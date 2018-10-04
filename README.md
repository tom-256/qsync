# qsync
 
## :pushpin: 概要
 QiitaのCLIクライアントです。  
以下の機能を使用することができます。
 - 投稿をローカルに保存
 - 投稿の更新
 - 新規投稿

## :globe_with_meridians: 動作環境
動作検証環境
```
 go version go1.10.3 darwin/amd64
```
 
## :floppy_disk: インストール方法
```
$go get github.com/ma-bo-do-fu/qsync
```
 
## :arrow_forward: 使用方法
### アクセストークンの生成
Qiitaの管理画面にアクセスし、アクセストークンを発行します  
https://qiita.com/settings/applications  
このとき、読み込み権限(read_qiita)と書き込み権限(write_qiita)を付与します。

### 設定ファイルの設置
設定ファイルを`~/.config/qsync/config.yaml`に設置します。
以下のようなyamlを書きます。
```
user_name: Qiitaのユーザ名
access_token: アクセストークン
local_root: 記事を保存するパス
```
### 投稿一覧の取得
```
$qsync pull
```
投稿一覧を取得します。  
以下のようなディレクトリ構造になります。
```
local_root/
├── 2017
│   └── 12
│       └── 05
│           └── 投稿のID.md
└── 2018
    ├── 01
    │   └── 17
    │       └── 投稿のID.md
    └── 10
        └── 01
            ├── 投稿のID.md
            └── 投稿のID.md
```
#### ファイルのフォーマット
```
---
Title: タイトル
Tags:
- name: '#Go'
  versions: 
    - 1.11.0
    - 1.10.0
- name: '#gcp'
  versions: []
Date: 2017-12-13T19:21:04+09:00
Url: https://qiita.com/USERNAME/items/xxx
Id: xxx
private: false
---

本文

```


### 投稿の更新
```
$qsync push /path/to/entry
```
投稿を更新します。

### 新規投稿
```
$qsync post
title:Example
tags:Go:1.11,1.10 gcp
```
新規に投稿します。投稿は下書きとして投稿された後、ローカルに保存されます。  
コマンド実行後、タイトルとタグを入力します。  
タイトルは255文字以下、タグは５つ以下です。   
公開するときは、ローカルで本文を書き、`private: false`として`qsync push`します。  
