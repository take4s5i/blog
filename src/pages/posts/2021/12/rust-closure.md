---
layout: ../../../../layouts/BlogPost.astro
title: Rustのクロージャを理解する
publishDate: 16 Dec 2021
---

まずはRustのクロージャを見てみましょう。

```rust
fn main() {
    let nums: Vec<i32> = vec![1,2,3,4,5];
    let y: i32 = 2;
    let iter = nums
        .into_iter()
        .map(|x| x * y);

    for x in iter {
        println!("{}", x);
    }
}
```

出力:
```
2
4
6
8
10
```

`|x| x * y` の部分がクロージャですね。
パット見単純そうに見えますが、他の言語のクロージャと同じように使おうとするとRustのライフタイムや所有権で躓きます。

例えば以下のコード。
クロージャの入門としてよく使われている？呼ばれた回数をカウントする関数ですが、これはコンパイルできません。

```rust
fn counter () -> Fn() -> i32 {
    let mut x: i32 = 0;
    return || {
        x += 1;
        x
    }
}

fn main() {
    let c = counter();
    println!("{}", c());
    println!("{}", c());
    println!("{}", c());
}
```

## Rustのクロージャとは何なのか
Rustのクロージャとは**特定のtraitを実装した構造体**です。
先ほどのコードは以下のような構造体がRustコンパイラによって生成されていると考えることができます（実際には動かないコードです)

```rust
struct MyClosure {
    x: i32
}

impl (Fn() -> i32) for MyClosure {
    call(&self) -> i32 {
        self.x += 1;
        x
    }
}

fn counter () -> impl Fn() -> i32 {
    let mut x: i32 = 0;
    return  MyClosure { x };
}

fn main() {
    let c = counter();
    println!("{}", c.call());
    println!("{}", c.call());
    println!("{}", c.call());
}
```

これで先ほどのコードがコンパイルできなかった理由がわかりやすくなりました。
- `self.x += 1` としているが、`self`がミュータブルではない
- `call`の型を`&mut self`にするなら`c`も`let mut c`でなければならない。

先ほどのコードを動くように直すとこのようになります。
```rust
fn counter () -> impl FnMut() -> i32 {
    let mut x: i32 = 0;
    return move || {
        x += 1;
        x
    }
}

fn main() {
    let mut c = counter();
    println!("{}", c());
    println!("{}", c());
    println!("{}", c());
}
```

## クロージャトレイト
Rustのクロージャは3種類あります。
- [Fn](https://doc.rust-lang.org/std/ops/trait.Fn.html)
- [FnMut](https://doc.rust-lang.org/std/ops/trait.FnMut.html)
- [FnOnce](https://doc.rust-lang.org/std/ops/trait.FnOnce.html)

トレイトの定義をこのようになっていますが、注目すべきは`self`の型です。

```rust
pub trait Fn<Args>: FnMut<Args> {
    extern "rust-call" fn call(&self, args: Args) -> Self::Output;
}

pub trait FnMut<Args>: FnOnce<Args> {
    extern "rust-call" fn call_mut(
        &mut self,
        args: Args
    ) -> Self::Output;
}

pub trait FnOnce<Args> {
    type Output;
    extern "rust-call" fn call_once(self, args: Args) -> Self::Output;
}
```

- `Fn`がイミュータブル参照`&self`
- `FnMut`がミュータブル参照`&mut self`
- `FnOnce`が所有権`self`
を取ることがわかります。

クロージャを構造体として考えると`FnOnce`がなぜ１度しか呼べないのかわかりやすいですね。

## クロージャのcaptureと環境
先ほどのこのコード
```rust
fn counter () -> impl FnMut() -> i32 {
    let mut x: i32 = 0;
    return move || {
        x += 1;
        x
    }
}
```

`x`はクロージャの中で参照されていますね。これをcaptureといいます。
なので`x`はクロージャにcaptureされていると言えます。

次にクロージャの構造体に相当する部分ですが、これをクロージャの`環境`と言ったりします。

変数がcaptureされるときはRustコンパイラが型を推測します。

[ここ](https://doc.rust-lang.org/reference/types/closure.html#capture-modes)に書いてあるように
- イミュータブル参照
- ミュータブル参照
- 所有権
の優先順になっているようです。

先ほどの例でいうと、ミュータブル参照としてcaptureしようとするとクロージャのライフタイムが`x`のライフタイムより長くなってしまいます。
そのため`move`キーワードを使ってRustコンパイラに所有権を渡すようにさせています。

## クロージャを引数に取る場合
- ジェネリック
- implトレイト
- dynトレイト
の３つがあります

```rust
fn closure_generic<T: Fn(i32) -> i32>(f: T) {
    println!("generic {}", f(1));
}

fn closure_impl(f: impl Fn(i32) -> i32) {
    println!("impl {}", f(1));
}

fn closure_dyn(f: Box<dyn Fn(i32) -> i32>) {
    println!("dyn {}", f(1));
}


fn main() {
    closure_generic(|x| x + 1);
    closure_impl(|x| x + 1);
    closure_dyn(Box::new(|x| x +1 ));
}
```

ジェネリックとimplトレイトについては引数に取る場合ほぼ同じです。(implトレイトのほうがシンプルでよさそうです。)
implとdynについては静的ディスパッチするか動的ディスパッチするかの違いになります。
dynの場合はトレイトオブジェクトを使っているのでBoxする必要があります。

## クロージャを返す場合
- implトレイト
- dynトレイト
の２つがあります

```rust
fn closure_impl() -> impl Fn(i32) -> i32 {
    return |x| x + 1;
}

fn closure_dyn() -> Box<dyn Fn(i32) -> i32 >{
    return Box::new(|x| x + 1);
}

fn main() {
    println!("impl {}", closure_impl()(1));
    println!("dyn {}", closure_dyn()(1));
}
```

クロージャの型はRustコンパイラが生成するのでプログラマが知ることはできません。
なのでジェネリックは使えずこの２パターンになります。

静的ディスパッチ、動的ディスパッチは引数で渡すときと同じですが、戻り値で返す場合`dyn`じゃないと扱えない場合があります。

```rust
fn closure_only_dyn(flg: bool) -> Box<dyn Fn(i32) -> i32> {
    if flg { Box::new(|x| x + 1) } else { Box::new(|x| x + 2) }
}
```

この関数のように条件によって異なるクロージャを返す場合です。
`|x| x + 1`と`|x| x + 2`はシグネチャが同じで実装しているトレイトも同じですが、あくまで別の型として扱われます。
そのためトレイトオブジェクトをつかった動的ディスパッチにする必要があります。

## まとめ
最初はよくわからなかったRustのクロージャですが、構造体であるということがわかったらだいぶ理解がはかどりました。
参考になれば幸いです。
