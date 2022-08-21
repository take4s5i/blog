---
layout: src/layouts/BlogPost.astro
title: GitHub Pagesで作ってるブログ編集リンクを追加する
publishDate: 11 Jul 2022
---

ブログに編集できるリンクを追加しました。
右上に出ている `Edit This Page` でGitHubの対応するmarkdownファイルに移動します。

## GitHubの編集リンク
GitHubの編集リンクを作るのは簡単で以下のようにすれば良いです。

```
https://github.com/{user or organization}/{repository}/blob/{branch}/{path to file}
```

このページの場合は
- user: `take4s5i`
- repository: `blog`
- branch: `main`
- path to file: `src/pages/posts/2022/07/adding-edit-link-to-posts.md`

なのでこのようなURLになります

```
https://github.com/take4s5i/blog/blob/main/src/pages/posts/2022/07/adding-edit-link-to-posts.md
```

このリンクで簡単にGitHub上の対応するファイルに飛んで編集できるというわけです。

`blob`のところを`edit`にすれば直接編集画面に飛べます

## AstroでPathを取得する方法
`Astro.request.url`で現在のURLを取得出来ます。

こんな感じ
```js
const url = new URL(Astro.request.url);
const removeTrailingSlash = (s) => s.endsWith('/') ? s.substring(0, s.length - 1) : s
const editUrl = `https://github.com/take4s5i/blog/blob/main/src/pages${removeTrailingSlash(url.pathname)}.md`
```

このURLを BlogPostのコンポーネントに埋め込めば完成です。
