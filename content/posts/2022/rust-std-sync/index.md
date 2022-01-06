---
title: "[Rust] 並列処理に使うAPIたち(Mutex, RwLock, Barrier, Condvar)"
slug: rust-std-sync
date: 2022-01-06
draft: false
tags:
  - rust
---
最近業務でgoで並列処理を書くことがあり、Rustの場合はどうなっているのか気になったので調べてみました。

[std::sync](https://doc.rust-lang.org/stable/std/sync/index.html)モジュールの中をいろいろ見ていきたいと思います。

# Mutex
まずは基本の[Mutex](https://doc.rust-lang.org/stable/std/sync/struct.Mutex.html)から。

RustのMutexは単にロックを制御するだけではなく、ロックで守られたデータを持つことができます。
（データが不要で単にロックだけほしい場合はユニット型`()`を使えばよいですね）

Mutexでカウンタをつくり、 4スレッド起動し、各スレッドで100回インクリメントを行います。
`cnt`は400になるはずです。

```rust
use std::sync::{Arc, Mutex, MutexGuard};
use std::thread;

const N_THREADS: usize = 4;

fn main() {
    let mut handles: Vec<thread::JoinHandle<_>> = Vec::with_capacity(N_THREADS);
    let mutex: Arc<Mutex<usize>> = Arc::new(Mutex::new(0));

    for _ in 0..N_THREADS {
        let mutex = mutex.clone();
        let handle = thread::spawn(move || {
            for _ in 0..100 {
                let mut cnt: MutexGuard<_> = mutex.lock().unwrap();
                *cnt += 1;
            }
        });

        handles.push(handle);
    }

    while let Some(handle) = handles.pop() {
        handle.join().unwrap();
    }

    let cnt = mutex.lock().unwrap();
    println!("cnt = {}", *cnt);
}
```

出力:
```
cnt = 400
```

ちゃんと400になりましたね。

まずMutexの生成ですが、通常、Mutexは複数スレッドで共有したいので`Arc<T>`を使います。
Mutex単体では`clone()`できませんが`Arc<T>`で包むことで複数スレッドでの共同所有を実現しています。
（`Arc<T>`は`Rc<T>`のスレッドセーフ版で、複数スレッドでの参照カウント方式のデータの共同所有ができます。そういえばこいつも`std::sync`ですね）

ロックの取得ですが`Mutex::lock()`を使います。
こいつは`LockResult<MutexGuard<T>>`という型のデータを返します。
`LockResult<T>` は`Result<T, PoisonError<T>>`という型のエイリアスです。

ロックを獲得した状態でそのスレッドがpanicすると、MutexがPoisoning状態になり、`lock()`がエラーを返すようになります。

無事ロックを獲得できると`MutexGuard<T>`が手に入りますが、こいつのライフタイムがそのままロックの取得期間になります。
つまり、スコープを抜けてdropされるとロックが解放されます。

Mutexで守られたデータのアクセスはこの`MutexGuard<T>`を使って行います。
可変参照を取得してデータを書き換えたり、不変参照を取って値を読み取ったりできます。

# RwLock
[RwLock](https://doc.rust-lang.org/stable/std/sync/struct.RwLock.html)はMutexに似たロック機構を提供しています。
- 読み取り専用の共有ロック
- 読み書きできる排他ロック

共有ロック取得中に排他ロックを取得したり、その逆はできません。
共有ロックは複数取得することができます。
(不変参照と可変参照のルールと同じですね）

メインスレッドで500msec毎にカウンタをインクリメントしていき、子スレッドでカウントが3より大きくなるまで待つ、ということをやってみようと思います。
```rust
use std::sync::{Arc, RwLock, RwLockReadGuard, RwLockWriteGuard};
use std::thread;

const N_THREADS: usize = 4;

fn main() {
    let mut handles: Vec<thread::JoinHandle<_>> = Vec::with_capacity(N_THREADS);
    let rwlock: Arc<RwLock<usize>> = Arc::new(RwLock::new(0));

    for n in 0..N_THREADS {
        let rwlock = rwlock.clone();
        let handle = thread::spawn(move || loop {
            {
                let guard: RwLockReadGuard<_> = rwlock.read().unwrap();
                println!("thread#{} get read lock", n);
                if *guard > 3 {
                    println!("thread#{} is breaking", n);
                    break;
                }
                println!("thread#{} drop read lock", n);
            }
            thread::sleep(std::time::Duration::from_secs(1));
        });

        handles.push(handle);
    }

    loop {
        {
            let mut guard: RwLockWriteGuard<_> = rwlock.write().unwrap();
            println!("main thread get write lock");
            *guard += 1;

            if *guard > 3 {
                println!("main thread is breaking");
                break;
            }
            println!("main thread drop write lock");
        }

        thread::sleep(std::time::Duration::from_millis(500));
    }

    while let Some(handle) = handles.pop() {
        handle.join().unwrap();
    }
}
```

出力:
```
thread#0 get read lock
thread#0 drop read lock
thread#1 get read lock
thread#1 drop read lock
thread#2 get read lock
thread#2 drop read lock
main thread get write lock
main thread drop write lock
thread#3 get read lock
thread#3 drop read lock
main thread get write lock
main thread drop write lock
thread#0 get read lock
thread#0 drop read lock
thread#1 get read lock
thread#2 get read lock
thread#1 drop read lock
thread#3 get read lock
thread#3 drop read lock
thread#2 drop read lock
main thread get write lock
main thread drop write lock
main thread get write lock
main thread is breaking
thread#0 get read lock
thread#0 is breaking
thread#1 get read lock
thread#1 is breaking
thread#3 get read lock
thread#3 is breaking
thread#2 get read lock
thread#2 is breaking
```

ちょっと長いですが、よくよく見てもらうとwriteロックは必ず排他的にかかるのに対してreadロックは複数同時取得している箇所があります。
readロックは`rwlock.read()`で、writeロックは`rwlock.write()`で取得できます。

Guardが返る点や使い方はMutexと同じです。

# Barrier
[Barrier](https://doc.rust-lang.org/stable/std/sync/struct.Barrier.html)は複数スレッド間でのタイミングの同期に利用できます。

`Barrier::new(n)`でBarrierを生成し、`n - 1`回目の`wait()`は呼び出し元スレッドをブロックします。
`n`回目の`wait()`を呼び出すとすべてのブロックが解除されタイミングが同期できるようになっています。

```rust
use std::sync::{Arc, Barrier};
use std::thread;

const N_THREADS: usize = 4;

fn main() {
    let mut handles = Vec::with_capacity(N_THREADS);
    let barrier: Arc<Barrier> = Arc::new(Barrier::new(N_THREADS));

    for n in 0..N_THREADS {
        let barrier = barrier.clone();
        let handle = thread::spawn(move || {
            println!("thead#{} is waiting", n);
            barrier.wait();
            println!("thead#{} is completed", n);
        });
        handles.push(handle);

        thread::sleep(std::time::Duration::from_secs(1));
    }

    while let Some(handle) = handles.pop() {
        handle.join().unwrap();
    }
}
```

出力:
```
thead#0 is waiting
thead#1 is waiting
thead#2 is waiting
thead#3 is waiting
thead#3 is completed
thead#2 is completed
thead#0 is completed
thead#1 is completed
```

# Condvar
[Condvar](https://doc.rust-lang.org/stable/std/sync/struct.Condvar.html)はスレッド間での通知に利用する条件変数というものです。
Mutexと一緒に利用します。

RwLockの例では各スレッドで`thread::sleep()`しながらループすることで待っていました。
これはビジーウェイトやビジーループと呼ばれるパターンで、効率的ではありません。
(CPUの時間を浪費しますし、sleepしている間は順番が回ってきても処理できないので非効率です)

Condvarを使ってビジーウェイトを回避しつつ他スレッドを待って見ます。
子スレッドでインクリメントしつつ、メインスレッドで10回カウントされるまで待ちます。

```rust
use std::sync::{Arc, Condvar, Mutex};
use std::thread;

const N_THREADS: usize = 4;

fn main() {
    let mut handles = Vec::with_capacity(N_THREADS);
    let pair = Arc::new((Mutex::new(0 as usize), Condvar::new()));

    {
        let pair = pair.clone();
        let handle = thread::spawn(move || {
            let (mutex, cvar) = &*pair;
            for n in 0..10 {
                {
                    let mut cnt = mutex.lock().unwrap();
                    *cnt += 1;
                    println!("child thread incrementes. cnt = {}", *cnt);
                }

                if n % 2 == 0 {
                    cvar.notify_one();
                }

                thread::sleep(std::time::Duration::from_millis(500));
            }
            cvar.notify_one();
        });
        handles.push(handle);
    }

    let (mutex, cvar) = &*pair;
    let mut cnt = mutex.lock().unwrap();
    while *cnt < 10 {
        println!("main thread waiting. cnt = {}", *cnt);
        cnt = cvar.wait(cnt).unwrap();
    }

    while let Some(handle) = handles.pop() {
        handle.join().unwrap();
    }
}
```

出力:
```
main thread waiting. cnt = 0
child thread incrementes. cnt = 1
main thread waiting. cnt = 1
child thread incrementes. cnt = 2
child thread incrementes. cnt = 3
main thread waiting. cnt = 3
child thread incrementes. cnt = 4
child thread incrementes. cnt = 5
main thread waiting. cnt = 5
child thread incrementes. cnt = 6
child thread incrementes. cnt = 7
main thread waiting. cnt = 7
child thread incrementes. cnt = 8
child thread incrementes. cnt = 9
main thread waiting. cnt = 9
child thread incrementes. cnt = 10
```

`cvar.notify_one()`されるたびに`cvar.wait()`しているメインスレッドが起動しカウンタの数をチェックしています。
（サンプルのためわざと途中で起動していますが、本来は不要かと）

`cvar.wait()`は`MutexGuard`の所有権を取り、所有権を返します。
waitするとnotifyされるまで呼び出しをブロックしつつ、ロックを解除します。

ぱっと見最初の`cnt`のロックが保持されたままになっているように見えますが、waitでいったん解除されるため子スレッドをブロックしてデッドロックになったりはしません。
notifyされてメインスレッドに処理が戻ってくると再びMutexGuardを取得しロックがかかります。

# 終わりに
[Once](https://doc.rust-lang.org/stable/std/sync/struct.Once.html)、[std::sync::atomic](https://doc.rust-lang.org/std/sync/atomic/)、
[std::sync::mpsc](https://doc.rust-lang.org/std/sync/mpsc/index.html)等ものちのち見ていこうかと思います。

