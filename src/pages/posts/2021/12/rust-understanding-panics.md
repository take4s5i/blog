---
layout: ../../../../layouts/BlogPost.astro
title: Rust panic を理解する
publishDate: 7 Dec 2021
---
panicしたら異常終了するぐらいにしか理解していなかったので。
いろいろ調べたり実験したりしてみました。

## panic したときの挙動

```rust
#[derive(Debug)]
struct Data(u32);

impl Drop for Data {
    fn drop(&mut self){
        println!("drop {:?}", self);
    }
}

fn call_recurse(n: u32) {
    if n == 0 {
        panic!("panic!");
    }
    let data = Data(n);
    call_recurse(n - 1);
    println!("return: {:?}", data);
}

fn main() {
    call_recurse(4);
}
```

出力:
```
  Compiling playground v0.0.1 (/playground)
    Finished dev [unoptimized + debuginfo] target(s) in 1.07s
     Running `target/debug/playground`
thread 'main' panicked at 'panic!', src/main.rs:12:9
note: run with `RUST_BACKTRACE=1` environment variable to display a backtrace

drop Data(1)
drop Data(2)
drop Data(3)
drop Data(4)
```

`Drop` トレイトを実装した構造体の生存中に`panic`したところ、ちゃんとdropされました。
この挙動は`unwinding`というようです。

## unwind と abort
panic時の挙動ですが、Cargoを使って変更できるようです。

[Profiles - The Cargo Book](https://doc.rust-lang.org/cargo/reference/profiles.html#panic)

- `panic` を`unwind`にすると`unwinding`を行う
- `panic` を`abort`にするとプロセスをその場で以上終了する。

`unwind`するとスレッドがパニックしてもハンドリングできるためプログラムは頑強になりそうです。
が、unwinding用のコードが各関数に埋め込まれるためコードサイズも大きくなるようなので一長一短ですね。

デーモンやWebサーバのといった長生きするプログラム以外では `abort` にしてしまってもいいかもしれません。

## panicのハンドリング
### set_hook
```rust
use std::panic;
use std::thread;

fn main() {
    std::panic::set_hook(Box::new(|_| {
        println!("Custom panic hook");
    }));
    let handle = thread::spawn(|| {
        panic!();
    });

    match handle.join() {
        Err(_) => println!("thread paniced"),
        Ok(_) => println!("thread exit successfly"),
    }

    panic!();
}
```

出力:
```
Custom panic hook
thread paniced
Custom panic hook
```

`set_hook` を使ってpanic時の処理をクロージャで渡します。
panicが起きると何度も呼ばれるようです。
このコードだとspawnした子スレッドのpanicとメインスレッドのpanicで２回呼ばれています。

`set_hook` で登録したパニックハンドラは`take_hook`という関数で登録解除できます。

### Thread
さっきしれっとやってましたがThreadもpanicをハンドリングする方法の１つです。
子スレッドでパニックすると親スレッド側で`join`したときに`Err`が返ります。

### catch_unwind
```rust
use std::panic;

fn main() {
    let result = panic::catch_unwind(|| {
        println!("hi!");
        panic!();
        println!("bye!");
    });

    match result {
        Err(_) => println!("paniced"),
        Ok(_) => println!("successfly"),
    }
}
```

出力:
```
hi!
paniced
```

`catch_unwind`を使うとスレッドを使わずにpanicを捕捉できます。

# 参考
- [Unrecoverable Errors with panic! - The Rust Programming Language](https://doc.rust-lang.org/book/ch09-01-unrecoverable-errors-with-panic.html)
- [Exception Safety - The Rustonomicon](https://doc.rust-lang.org/nomicon/exception-safety.html)
- [std::panic - Rust](https://doc.rust-lang.org/std/panic/index.html)
- [set_hook in std::panic - Rust](https://doc.rust-lang.org/std/panic/fn.set_hook.html)
- [abort in std::process - Rust](https://doc.rust-lang.org/std/process/fn.abort.html)
- [(4) Unwinding vs Abortion upon panic : rust](https://www.reddit.com/r/rust/comments/phws7n/unwinding_vs_abortion_upon_panic/hbncri9/?utm_source=share&utm_medium=web2x&context=3)
