---
title: "[Rust] 並列処理に使うAPIたち2(Once, channel)"
slug: rust-std-sync2
date: 2022-01-29
draft: false
tags:
  - rust
---

[前回](https://blog.take4s5i.dev/posts/2022/rust-std-sync/)の続きで Once と mpsc を調べていきます。

# Once

[Once](https://doc.rust-lang.org/stable/std/sync/struct.Once.html) は static 変数の初期化に利用できます。

Once を使うことで、複数スレッドから同時に初期処理が実行されても、一度だけ実行されることを保証できます。

```rust
use std::collections::HashMap;
use std::sync::Once;
use std::thread;

static mut CONFIG: Option<HashMap<String, String>> = None;

static START: Once = Once::new();

fn get_config(key: &str) -> Option<String> {
    unsafe {
        START.call_once(|| {
            println!("call_once");
            CONFIG = Some(HashMap::new());

            CONFIG
                .as_mut()
                .unwrap()
                .insert("hoge".to_owned(), "12345".to_owned());
        });
        CONFIG.as_ref().unwrap().get(key).map(|v| v.clone())
    }
}

fn main() {
    let mut handles = Vec::new();
    for i in 0..10 {
        handles.push(thread::spawn(move || {
            println!("thread#{} {}", i, get_config("hoge").unwrap());
        }));
    }

    while let Some(h) = handles.pop() {
        h.join().unwrap()
    }
}
```

出力:

```
call_once
thread#0 12345
thread#4 12345
thread#3 12345
thread#5 12345
thread#1 12345
thread#6 12345
thread#7 12345
thread#2 12345
thread#8 12345
thread#9 12345
```

複数スレッドから`call_once`が実行されていますが、一度しか初期化処理が実行されていないことがわかります。

また`call_once`は初期化処理中に別スレッドから呼び出されると、初期化が終わるまでブロックするため、
初期化されていない状態で変数にアクセスしてしまうこともありません。

mut な static 変数にアクセスするには`unsafe`を使う必要があるため、関数にするなどしてラップしてあげるのが良さそうです。

std だけでも static なグローバル変数は実現できますが、素直に[once_cell](https://crates.io/crates/once_cell)を使った方が使いやすそうです。

# channel

[std::mpsc::channel](https://doc.rust-lang.org/stable/std/sync/mpsc/fn.channel.html) はスレッド間での通信に利用できる FIFO キューを作る関数です。

モジュール名の`mpsc`は multiple producer single consumer の略で、
producer（キューにデータを入れる側）は複数、consumer（キューからデータを取り出す側）は 1 つということを表しています。

複数の子スレッドからデータを送信し、メインスレッドで受信するサンプルを作ってみます。

```rust
use std::sync::mpsc::channel;
use std::thread;

fn main() {
    let mut handles = Vec::new();
    let (sender, receiver) = channel::<String>();

    for n in 0..4 {
        let sender = sender.clone();
        handles.push(thread::spawn(move || {
            for i in 0..4 {
                sender
                    .send(format!("thread#{} val = {}", n, i).to_owned())
                    .unwrap();
            }
        }));
    }

    drop(sender);

    while let Ok(v) = receiver.recv() {
        println!("{}", v);
    }

    while let Some(h) = handles.pop() {
        h.join().unwrap()
    }
}
```

出力:

```
thread#2 val = 0
thread#3 val = 0
thread#1 val = 0
thread#3 val = 1
thread#1 val = 1
thread#3 val = 2
thread#1 val = 2
thread#3 val = 3
thread#1 val = 3
thread#0 val = 0
thread#0 val = 1
thread#0 val = 2
thread#0 val = 3
thread#2 val = 1
thread#2 val = 2
thread#2 val = 3
```

`channel`関数は Sender と Receiver のタプルを返します。

これはお互いに関連づけられていて、Sender から送られたデータは対にになる Receiver で受信できます。

また、Sender は`clone`できるので clone したものを別スレッドに move します。

`sender.send(val)`でデータを送信し `receiver.revc()`でデータを受信しますがこれらは以下のケースで Err を返します。

- `sender.send(val)`: 対応する`receiver`が drop されていた場合
- `receiver.recv()`: 対応する全ての`sender`が drop されていた場合

メインスレッド側 `drop(sender)` していますが、これをしないとメインスレッドの`sender`が drop されるのを待ち続けてデッドロックになります。
