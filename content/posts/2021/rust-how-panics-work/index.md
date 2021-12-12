---
title: "[Rust] panic の仕組み"
date: 2021-12-07
draft: false
---
# [Rust] panic の仕組み
Rust には複数のエラーハンドリング方法が存在しています。

- `Result`, `Option`
- `panic`
- `abort`

基本的には `Result` や `Option` を用いるべきです。
これらは回復可能なエラーを表していてパターンマッチを用いてユーザがエラーハンドリング処理を実装できます。

`panic` と `abort` は回復不能なエラーを表していて、発生するとプログラムが異常終了します。

## unwinding on panic
`panic` を起こすとすべてを中断して即座にプロセスを終了するような印象があるかもしれません。
（自分もそうでしたが)

`panic` を起こしたときに終了するのはスレッドです。
スレッドで `panic` を起こすと `join()` したときに `Err` が返りますので親スレッドは子スレッドが `panic` してもハンドリングすることができます。

メインスレッドが `panic` を起こすとその時はプロセスが終了します。

スレッドが終了するとき、コールスタック上に配置されているローカル変数 drop する必要があります。
(dropしないとメモリリークです)

この処理を`unwinding`といい、Rustコンパイラが勝手に埋め込んでくれます。
これはJava等の例外処理における`catch`に似た挙動ですが、ユーザが自分でcatchの部分にコードを書くことはできません。

`unwinding` は安全性を高めてくれていますが、いいことばかりではないようです。
- `unwinding`用のコードが各関数に埋め込まれる = バイナリサイズの肥大化
- `unwinding`処理自体が遅くなったり、panicする可能性がある

これに対し`abort` は`unwinding`を行いません。
つまり、安全性は担保されていませんが、軽量ということです。
（プログラム事終了すれば良い場合は`abort`でよさそうな気もしますね)

## panicを処理する方法
Rustの `panic` は例外のようにcatchしてリカバリすることはできません。
一度おこると回復できません。

じゃあ `panic` したときに何ができるかというと。
- `drop`: これはpanicとかに関係なくちゃんと実装しておけって話ですね
- 親スレッドでのエラーハンドリング
  - 子スレッドの `panic` は親スレッドで`Result`として処理できます。
- [std::panic::catch_unwind](https://doc.rust-lang.org/std/panic/fn.catch_unwind.html)をつかう
  - スレッドを作らずに`panic`を`Result`として扱えます。
  - closureの中で `panic` すると `Result` が `Err` になるようです
- [std::panic::set_hook](https://doc.rust-lang.org/std/panic/fn.set_hook.html)
  - panicがおこったときにclosureを呼んでくれるようです。
  - 回復できるわけではないのでログに書き込んだりするのに使えそうです。

# 参考
- [Unrecoverable Errors with panic! - The Rust Programming Language](https://doc.rust-lang.org/book/ch09-01-unrecoverable-errors-with-panic.html)
- [Exception Safety - The Rustonomicon](https://doc.rust-lang.org/nomicon/exception-safety.html)
- [std::panic - Rust](https://doc.rust-lang.org/std/panic/index.html)
- [set_hook in std::panic - Rust](https://doc.rust-lang.org/std/panic/fn.set_hook.html)
- [abort in std::process - Rust](https://doc.rust-lang.org/std/process/fn.abort.html)
- [(4) Unwinding vs Abortion upon panic : rust](https://www.reddit.com/r/rust/comments/phws7n/unwinding_vs_abortion_upon_panic/hbncri9/?utm_source=share&utm_medium=web2x&context=3)

