---
layout: src/layouts/BlogPost.astro
title: ブログをAstroに移行しました
publishDate: 9 Jul 2022
---

もともとブログを [Hugo](https://gohugo.io/) で作っていましたが
[Astro](https://astro.build/) に移行しました。

Hugo は様々なテーマがあって簡単に使えるところに魅力を感じていましたが、
テーマ毎に使い方を覚えなくてはならず、ドキュメントが不十分なテーマもあったため使いづらさを感じていました。
（goテンプレートにあまり馴染みがないというのもあります）

React や JSX と言った昨今のフロントエンド技術スタックのほうが馴染みはあるんですが、
ブログのためだけに Next.js で SSG してブログの仕組みを作るのは面倒くさくてやっていませんでした。

最近 Astro を知り、ブログ用途にも簡単に使えそうだったので切り替えることにしました。

## Astroでブログを作る
`yarn create astro` を実行して質問に答えただけで簡単にブログを作ることができました

```sh
❯ yarn create astro
yarn create v1.22.19
[1/4] 🔍  Resolving packages...
[2/4] 🚚  Fetching packages...
[3/4] 🔗  Linking dependencies...
[4/4] 🔨  Building fresh packages...

success Installed "create-astro@0.12.4" with binaries:
      - create-astro
[####################################################] 52/52
Welcome to Astro! (create-astro v0.12.4)
Lets walk through setting up your new Astro project.

✔ Where would you like to create your new project? … ./my-astro-site
✔ Which template would you like to use? › Blog
✔ Template copied!
✔ Would you like us to run "yarn install?" … yes
✔ Packages installed!
✔ Initialize a new git repository? This can be useful to track changes. … yes
✔ Setup complete.
✔ Ready for liftoff!

 Next steps

You can now cd into the my-astro-site project directory.
Run yarn dev to start the Astro dev server. CTRL-C to close.
Add frameworks like react and tailwind to your project using astro add

Stuck? Come join us at https://astro.build/chat
Good luck out there, astronaut.
```

## 記事の移行
Astro はビルトインで `.md` ファイルを読み込めるようになっていて、
Syntax Highlight 含め Markdown を使うためになにか特別な設定は必要ありませんでした。

frontmatter は Hugo と同じく yaml だったので、項目の手直し程度ですみました。

ブログを生成すると `src/pages/posts/index.md` にサンプルの記事が生成されるのでそれを元に
frontmatterを直しました。

frontmatterで `layout` を指定しないと崩れます。

```yaml
---
layout: path/to/Layout.astro
# example
# layout: ../../layouts/BlogPost.astro
---

contents ....
```

生成されたテンプレートでは `setup` で `import` 文を使用してレイアウトを読み込んでいましたが、最新のドキュメントでは `layout`になっているようです。
ここらへんはまだ beta なので仕方ないでしょう。

`title` とか `description` とか色々他に設定できるものはありましたが、
これらは結局 Layout ファイルの中で使われているだけなので、必要なものだけ使えば良さそうです。

Astro はコンポーネント志向の流れを汲んでいて、個人的にかなりとっつきやすかったです。

普通にjsでかけるのでカスタマイズもそんなに難しくないでしょう。
vimのハイライトがうまく効かないのが現状の難点ですが、まぁ記事を書く文にはそんなに問題なかったです。

## 記事の執筆, 公開
`astro dev` でファイルの変更を watch してリロードしてくれます。

公開するときは `astro build` で SSG 出来ます。

GitHub Pages はじめ、様々な環境にデプロイするための方法が公式ドキュメントで公開されていたので
コピペするだけで簡単に対応することが出来ました。

（今回使ったのは[これ](https://docs.astro.build/en/guides/deploy/github/)です）

このブログはカスタムドメインなので cname 設定しないと見れなくなってしまうので注意が必要でした。
