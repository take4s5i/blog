---
title: "Rustのcrossbeamを試してみる"
date: 2022-02-03
draft: false
tags:
  - rust
  - corssbeam
---
Rustのライブラリである[crossbeam](https://github.com/crossbeam-rs/crossbeam)を試してみます。

crossbeamは公式の説明によると

    This crate provides a set of tools for concurrent programming

このcrateは並列プログラミングのためのツールセットを提供する、とあります。

並列プログラミングを行う際のデータ構造や便利な関数が実装されているようです。
それではみていきます。

# crossbeam::scope
[crossbeam::scope](https://docs.rs/crossbeam/0.8.1/crossbeam/fn.scope.html)関数は
スコープ付きのスレッドを作ることができます。

スコープ付きというのはライフタイムが制限されているということです。
通常のスレッドはいつまでスレッドが生きるのかわからないため、スレッドの外から借用するには`'static`ライフタイムが必要でした。

スコープ付きスレッドを使うと、Rustコンパイラがスレッドのライフタイムを認識してくれるため、
`'static`出なくてもデータを借用できるようになります。

## std::threadでの例
まずはcrossbeamを使わない普通のstd::threadのコードをみていきます。
```rust
use std::{
    collections::HashMap,
    fs,
    sync::{Arc, Mutex},
};

fn main() {
    let files = vec!["a.txt", "b.txt", "c.txt"];
    let contents: Arc<Mutex<HashMap<String, String>>> = Arc::new(Mutex::new(HashMap::new()));

    let mut h = vec![];
    // --------- ここでエラー
    for file in files.iter() {
        let file = file;
        let contents = contents.clone();
        h.push(std::thread::spawn(move || {
            let s = fs::read_to_string(file).unwrap();
            let mut m = contents.lock().unwrap();
            m.insert(file.to_string(), s);
        }));
    }
    while let Some(h) = h.pop() {
        h.join().unwrap();
    }

    let contents = Mutex::into_inner(Arc::try_unwrap(contents).unwrap()).unwrap();

    for (f, c) in contents.iter() {
        println!("file: {}", f);
        println!("---------------");
        println!("{}", c);
        println!();
    }
}
```
これはコンパイルエラーになります。

```
error[E0597]: `files` does not live long enough
  --> src/main.rs:13:17
   |
13 |     for file in files.iter() {
   |                 ^^^^^^^^^^^^
   |                 |
   |                 borrowed value does not live long enough
   |                 argument requires that `files` is borrowed for `'static`
...
34 | }
   | - `files` dropped here while still borrowed
```

`files` からファイル名を借用してスレッドに送っていますが、
子スレッドの方が`files`より長生きする可能性があるとしてエラーになっています。
（実際のところは, `files`のライフタイム内で子スレッドをjoinしているため、起こり得ません。）

## crossbeamでの例
次にcrossbeamのscope関数を使ってみます。

```rust
use crossbeam::thread;
use std::{
    collections::HashMap,
    fs,
    sync::{Arc, Mutex},
};

fn main() {
    let files = vec!["a.txt", "b.txt", "c.txt"];
    let contents: Arc<Mutex<HashMap<String, String>>> = Arc::new(Mutex::new(HashMap::new()));

    thread::scope(|s| {
        for file in files.iter() {
            let file = file;
            let contents = contents.clone();
            s.spawn(move |_| {
                let s = fs::read_to_string(file).unwrap();
                let mut m = contents.lock().unwrap();
                m.insert(file.to_string(), s);
            });
        }
    })
    .unwrap();

    let contents = Mutex::into_inner(Arc::try_unwrap(contents).unwrap()).unwrap();

    for (f, c) in contents.iter() {
        println!("file: {}", f);
        println!("---------------");
        println!("{}", c);
        println!();
    }
}
```

出力:
```
file: b.txt
---------------
bbbbbb


file: a.txt
---------------
aaaaa


file: c.txt
---------------
ccccc
```

コンパイルが通り期待通り動作しています。

`s.spawn`で起動したスレッドで手動でjoinする必要はありません。
`scope`関数は`s.spawn`したスレッドが全て完了するまで、自動的にブロックします。

ソースを読んでみた感じだと、
[std::thread::Builder](https://doc.rust-lang.org/std/thread/struct.Builder.html)を使って
スレッドを起動させているようでした。

つまり、スコープ付きで扱えるだけで普通にOSのネイティブスレッドが起動しているため、
グリーンスレッドのようにカジュアルにスレッドを起動する用途には向いていないようです。

# crossbeam::channel
[crossbeam::channel](https://docs.rs/crossbeam/0.8.1/crossbeam/channel/index.html)は
スレッド間のメッセージ送受信に利用するライブラリです。

使い方は[std::mpsc::channel](https://blog.take4s5i.dev/posts/2022/rust-std-sync2/)とほとんど同じです。

違いとしては、
- stdの方は multi-producer, single-consumerに対して
- crossbeamの方は multi-producer, multi-consumer
になっていることが挙げられます。

また、公式の説明によるとstdより機能豊富でパフォーマンスが良いそうです。

チャネルを作るために２種類の関数が提供されており
- [bounded](https://docs.rs/crossbeam/0.8.1/crossbeam/channel/fn.bounded.html)
    - メッセージのバッファが有限
    - バッファに空きがない場合、`send`をブロックしてバッファが開くのを待つ
- [unbounded](https://docs.rs/crossbeam/0.8.1/crossbeam/channel/fn.unbounded.html)
    - メモリの許す限りメッセージをバッファすることができる

`bounded`を使ってサンプルを作ってみます。

```rust
use crossbeam::{channel, thread};
use std::fs;

fn main() {
    thread::scope(|s| {
        let (tx_file, rx_file) = channel::bounded::<String>(0);
        let (tx_content, rx_content) = channel::bounded::<(String, String)>(0);

        // sending file name
        s.spawn(move |_| {
            let files = vec!["a.txt", "b.txt", "c.txt"];
            for f in files.iter() {
                println!("[sender] send {}", f);
                tx_file.send(f.to_string()).unwrap();
                std::thread::sleep(std::time::Duration::from_millis(200));
            }
        });

        // reading files
        s.spawn(move |_| {
            for file in rx_file.into_iter() {
                println!("[reader] read {}", file);
                let content = fs::read_to_string(&file).unwrap();
                tx_content.send((file, content)).unwrap();
                std::thread::sleep(std::time::Duration::from_millis(500));
            }
        });

        // consuming results
        s.spawn(move |_| {
            for (f, c) in rx_content.into_iter() {
                println!("[consumer] {} = {:?}", f, c);
                std::thread::sleep(std::time::Duration::from_millis(1000));
            }
        });
    })
    .unwrap();
}
```

出力:
```
[sender] send a.txt
[reader] read a.txt
[consumer] a.txt = "aaaaa\n"
[sender] send b.txt
[reader] read b.txt
[sender] send c.txt
[consumer] b.txt = "bbbbbb\n"
[reader] read c.txt
[consumer] c.txt = "ccccc\n"
```

```rust
let (tx_file, rx_file) = channel::bounded::<String>(0);
let (tx_content, rx_content) = channel::bounded::<(String, String)>(0);
```

まず、スレッド間通信ようのチャネルを`bounded`を使って作ります。
キャパシティとして`0`を指定しているので、このチャネルにはバッファがありません。
そのため、`send`した場合は`recv`されるまでブロックされることになります。

後続処理の方が処理時間が長く`send`をブロックするため綺麗に順番に実行されています。

これを次のように`3`で実行してみます。
```rust
let (tx_file, rx_file) = channel::bounded::<String>(3);
let (tx_content, rx_content) = channel::bounded::<(String, String)>(3);
```

出力:
```
[sender] send a.txt
[reader] read a.txt
[consumer] a.txt = "aaaaa\n"
[sender] send b.txt
[sender] send c.txt
[reader] read b.txt
[consumer] b.txt = "bbbbbb\n"
[reader] read c.txt
[consumer] c.txt = "ccccc\n"
```

最初の`sender`, `reader`, `consumer`の順番は同じですが、`reader`が`recv`するよりも先に
`sernder`が`send`できていることがわかります。
