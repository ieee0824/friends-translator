# friends-translator

単純な決め打ちのネガポジ判定での生成

```
$ go run main.go -i 私は歌が好き
```

ナイーブベイズを使う

```
$ go run main.go -m -i 私は歌が好き
```

ベイズ分類器を学習させる

```
$ go run nb/main.go -i ほげほげ
```
