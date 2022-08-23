---
layout: src/layouts/BlogPost.astro
title: AWS CDK で同じ環境で複数回 bootstrap する
publishDate: 23 Aug 2022
---

## TL;DR
### cdk bootstrap の実行
```
cdk bootstrap --qualifier my-qualifier --toolkit-stack-name my-cdktoolkit-stack
```
- `my-qualifier` は他とかぶらない名前にする
- `my-cdktoolkit-stack` は他とかぶらない Stack 名にする

### qualifier の設定

**cdk.json**
```json
{
  "context": {
    "@aws-cdk/core:bootstrapQualifier": "my-qualifier"
  }
}
```

### IAM Userの設定
デプロイに利用する IAM User に対して `cdk-my-qualifier-*` という名前の IAM Role を Assume できるように設定する。

## cdk bootstrap ことはじめ
いつのまにやら `cdk bootstrap` が大幅進化していたので改めて理解します。
`cdk bootstrap` を実行すると、環境変数や `cdk.json` などから現在の環境を推測します

環境というのは AWS アカウントとリージョンのことで

```
cdk bootstrap aws://123456789012/ap-northeast-1
```

のように具体的に指定することも出来ます。

bootstrap が始まると環境で指定された AWS アカウント、リージョン内に `CDKToolkit` という Cloud Formation スタックが作成されます。

そのため、bootstrap を実行する IAM User/Role は `CDKToolkit` スタックを作れるだけのパーミションを持っている必要があります。

bootstrap に必要な IAM Policy が書いてないのは不親切感はありますが、`cdk bootstrap --show-template` を実行すると
`CDKToolkit` の Cloud Formation テンプレートが表示されるのでそこから必要最低限のパーミッションだけ抜き出す事ができるでしょう。

## `CDKToolkit` の中身
肝心の `CDKToolkit` の中身ですが以下のようなリソースが作成されていました。

|論理ID|物理ID|タイプ|
|---|---|---|
|CdkBootstrapVersion|/cdk-bootstrap/hnb659fds/version|AWS::SSM::Parameter|
|CloudFormationExecutionRole|cdk-hnb659fds-cfn-exec-role-123456789012-ap-northeast-1|AWS::IAM::Role|
|ContainerAssetsRepository|cdk-hnb659fds-container-assets-123456789012-ap-northeast-1|AWS::ECR::Repository|
|DeploymentActionRole|cdk-hnb659fds-deploy-role-123456789012-ap-northeast-1|AWS::IAM::Role|
|FilePublishingRole|cdk-hnb659fds-file-publishing-role-123456789012-ap-northeast-1|AWS::IAM::Role|
|FilePublishingRoleDefaultPolicy|CDKTo-File-8GKTZEZLLE17|AWS::IAM::Policy|
|ImagePublishingRole|cdk-hnb659fds-image-publishing-role-123456789012-ap-northeast-1|AWS::IAM::Role|
|ImagePublishingRoleDefaultPolicy|CDKTo-Imag-1IOQ0HFE6YPV9|AWS::IAM::Policy|
|LookupRole|cdk-hnb659fds-lookup-role-123456789012-ap-northeast-1|AWS::IAM::Role|
|StagingBucket|cdk-hnb659fds-assets-123456789012-ap-northeast-1|AWS::S3::Bucket|
|StagingBucketPolicy|CDKToolkit-StagingBucketPolicy-19C4MIKTPA6IG|AWS::S3::BucketPolicy|

物理IDに規則性が見えますね。

- `hnb659fds`: Qualifier (後述)
- `123456789012`: AWS Account ID
- `ap-northeast-1`: AWS Region

となっています。

### CdkBootstrapVersion
これはおそらく `CDKToolkit` に使われている Cloud Formation テンプレートのバージョンでしょう。
この記事を書いている時点で値は `14` でした。

### CloudFormationExecutionRole
CDK は最終的に Cloud Formation テンプレートを作成しそれをデプロイしますが、
このロールは CloudFormation を実行する際に使われると思われます。

デフォルトでは `AdministratorAccess` が付与されていて、非常に強い権限となっています。

cdk bootstrap のオプション `--cloudformation-execution-policies` を指定することで、
このロールに付与する権限を変更できるようです。

### StagingBucket, FilePublishingRole
CDK ではローカルにあるディレクトリやファイルをS3 バケットにアップロードする
[Assets](https://docs.aws.amazon.com/cdk/v2/guide/assets.html) という機能を持っています。

これは Lambda を使うときに非常に便利です。
Lambda を Cloud Formation でデプロイする場合、コードを zip で固めて s3 に事前にアップロードしておき、
その s3 パスを Cloud Formation テンプレートに指定する必要があります。

CDK ではその作業を自動でやってくれます。
おそらく、`cdk deploy` の中で、Cloud Formation のデプロイ前に assets を s3 バケットにアップロードするような処理を行っているのでしょう。

`StagingBucket` と `FilePublishingRole` はこのとき使われるアップロード先とアップロード用の IAM Role と思われます。

### ContainerAssetsRepository, ImagePublishingRole
CDK の Assets は S3 バケットにファイルをアップロードするのと同じ要領で
ECR へのコンテナイメージの push も出来ます。

同様の使い方でしょう。

### LookupRole
CDK では [Context](https://docs.aws.amazon.com/cdk/v2/guide/context.html) という機能で
デプロイ時に環境から vpc id などを取得することが出来ます。

`LookupRole` は Context 取得時に使われる IAM Role かと思われます。

### DeploymentActionRole
こいつだけよくわかりませんでした・・・
CLI や Pipeline でデプロイするのに使われると書いてあったので、`cdk deploy` したときにこのRole を Assume し、
そこから更に `LookupRole` などを Assume するようなイメージでしょうか

## cdk deploy
`cdk deploy` するときには `cdk bootstrap` で作られた IAM Role や S3 Bucket を使うわけですが、
一体どうやってその名前を取得しているのでしょうか？

`CDKToolkit` のスタックから物理IDを取得しているのかと思っていましたが、
どうやらそうではなく、物理IDの命名規則に従って特定しているようです。

この挙動は [DefaultStackSynthesizer](https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.DefaultStackSynthesizer.html)
を明示的に指定することで他の Role や S3 Bucket を使ってデプロイすることが出来ます。

また、Role を使ってデプロイすることになる都合上、`cdk deploy` を実行する
IAM User は bootstrap で作られた Role を Assume できる必要があります。

Assume 出来ない場合は IAM User の権限でそのまま実行されますが、Assume できるようにしておいたほうが楽だし無難でしょう。

## Qualifier
ここでやっと登場するのが `Qualifier` です。

`Qualifier` のデフォルト値は `hnb659fds` ですが、明示的に指定することが出来ます。
- `cdk bootstrap` の `--qualifier` オプション
- `cdk.json` の `context.@aws-cdk/core:bootstrapQualifier`
- `DefaultStackSynthesizer` の `qualifier` プロパティ

いつ qualifier を使うんだという話ですが、
`cdk bootstrap` は環境毎に行う都合上、同じ AWS アカウント、リージョン内で１回しかbootstrap出来ません。

これは以下のようなケースで使いにくい場合があるでしょう。
- 同じ環境を複数チームで使っている場合
    - CDK bootstrap バージョンが異なると壊れる/壊される可能性がある
- 同じ環境に複数の CDK App があり、それぞれ `--cloudformation-execution-policies` を分けたい場合
- たまたま同じ名前のリソースがあり、コンフリクトした場合
    - 前のcdk bootstrap の残骸とか
    - Staging Bucket は `CDKToolkit` を消しても残る

Qualifier を使うとリソース名のコンフリクトを回避出来ます。
この場合 `--toolkit-stack-name` で `CDKToolkit` のスタック名を変えておく必要があります。
（変えていないと 既存の `CDKToolkit` を上書きして破壊することになりそう）

また bootstrap で作られる IAM Role は `cdk-{qualifier}-*` で始まるので、
IAM User に `sts::AssumeRole` を設定するときにリソースを指定することで不要なRoleへのAssumeを拒否出来ます。

##
- [Bootstrapping - AWS Cloud Development Kit (AWS CDK) v2](https://docs.aws.amazon.com/cdk/v2/guide/bootstrapping.html)
- [class DefaultStackSynthesizer · AWS CDK](https://docs.aws.amazon.com/cdk/api/v2/docs/aws-cdk-lib.DefaultStackSynthesizer.html)

