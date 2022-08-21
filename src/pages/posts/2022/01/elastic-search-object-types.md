---
layout: src/layouts/BlogPost.astro
title: Elastic Searchのobject, flattened, nestedの違い
publishDate: 29 Jan 2022
---

Elastic Search の object 系の type について調べたのでまとめていきます。

# object

[object](https://www.elastic.co/guide/en/elasticsearch/reference/current/object.html)は JSON のオブジェクトなどの
データ構造をインデックスするのに使える型です。

```json
{
  "user": {
    "name": "tanaka",
    "age": 20
  }
}
```

このようなオブジェクトがあった場合

```json
{
  "user.name": "tanaka",
  "user.age": 20
}
```

のようにフラットにされてインデックスされます。

オブジェクトにすることでパフォーマンス的なデメリットは特になさそうだったので、
単純にフィールドを構造化して整理するのに利用すると良さそうです。

ただし、オブジェクトの配列を持つ場合はおそらく意図した挙動にならないので注意が必要です。

例えば以下のようなデータがあったとします

```json
{
  "skills": [
    { "name": "ElasticSearch", "level": "newbie" },
    { "name": "Typescript", "level": "advanced" }
  ]
}
```

これは次のようにインデックスされます。

```json
{
  "skills.name": ["ElasticSearch", "Typescript"],
  "skills.level": ["newbie", "advanced"]
}
```

フラットにされてしまうことで、name と level の関係性が失われてしまいます。

```json
{
  "query": {
    "bool": {
      "must": [
        { "term": { "skills.name": "ElasticSearch" } },
        { "term": { "skills.level": "advanced" } }
      ]
    }
  }
}
```

このクエリの意図は「ElasticSearch が advanced な人」ですが、うまく機能しません。
先ほどのデータでは、ElasticSearch は newbie だったためヒットするべきではありませんが、これがヒットしてしまいます。

# nested

これを解決するのが[nested](https://www.elastic.co/guide/en/elasticsearch/reference/current/nested.html)です。

`nested` type としてフィールドを定義すると、フィールド同士の関係性が維持されるようになります。

```json
{
  "mappings": {
    "properties": {
      "skills": {
        "type": "nested"
      }
    }
  }
}
```

これは内部的には別のドキュメントとしてインデックスすることで実現しているようです。
（skills を子テーブルのように持って join しているイメージ？）

クエリする際も[nested query](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-nested-query.html)という専用のクエリを利用する必要があります。

```json
{
  "query": {
    "nested": {
      "path": "skills",
      "query": {
        "bool": {
          "must": [
            { "term": { "skills.name": "ElasticSearch" } },
            { "term": { "skills.level": "advanced" } }
          ]
        }
      }
    }
  }
}
```

これでやっと意図通りに検索できます。

ただし、nested は join する特性上 ElasticSearch 的にかなり重いクエリのようなので、
検索のパフォーマンスを最大化するなら非正規化する等で避けた方が良さそうです。

# flattened

[flattened](https://www.elastic.co/guide/en/elasticsearch/reference/current/flattened.html)は少し特殊な object 型です。

基本的な挙動は object と似ていて、フィールド間の関係性が保持されない点も同じです。

object との違いは、子フィールドがそれぞれ別フィールドとして扱われるか、１つのフィールドとして扱われるか、という点です。

```json
{
  "skills": [
    { "name": "ElasticSearch", "level": "newbie" },
    { "name": "Typescript", "level": "advanced" }
  ]
}
```

このようなデータは flattened では次のようにインデックスされます。

```json
{
  "skills": [
    "name\u0000ElasticSearch",
    "level\u0000newbie",
    "name\u0000Typescript",
    "level\advanced"
}
```

`\u0000`は ElasticSearch 内部で利用される区切り文字です。
（[ソース](https://github.com/elastic/elasticsearch/blob/58ce0f94a0bbdf2576e0a00a62abe1854ee7fe2f/server/src/main/java/org/elasticsearch/index/mapper/flattened/FlattenedFieldParser.java#L31)を読んだ感じ、null 文字で区切られているようです。）

見ての通り、`skills`という 1 つのフィールドに値としてフラットにされています。
また、子フィールドの値が数値であっても文字列に返還されるのも特徴で、内部的には`keyword`型としてインデックスされています。

flattened はフィールド名が動的に変わったりする場合に使えそうです。

object にしてしまうと、フィールド名が変わったり追加するたびにフィールドが増えて mapping が肥大化していきますが、
flattened にすることで１つのフィールドとして扱えます。

# まとめ

- object はフィールドを階層化してグルーピングしたい場合に使える
- nested はオブジェクトの配列を関係性を維持したままインデックス、クエリしたい場合に使える
- flattened はフィールド名が動的に変化する場合に使える

# 参考

- https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html#object-types
- https://www.elastic.co/jp/blog/managing-relations-inside-elasticsearch
- https://opster.com/guides/elasticsearch/data-structuring/elasticsearch-nested-field-object-field/
