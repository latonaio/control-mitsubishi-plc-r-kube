# control-mitsubishi-plc-r-kube
三菱電機製のPLCのレジスタに登録されたメッセージを読み込むマイクロサービスです。

メッセージの送受信方法およびフォーマットはMCプロトコルに準じています。

## MCプロトコル
三菱電機製レジスタに採用されている、三菱電機独自のプロトコルです。

16進数のバイナリで構成された電文を送受信し、レジスタに対して操作を行うメッセージングプロトコルです。


[MCプロトコルのマニュアル（三菱電機のHPに遷移します）](https://www.mitsubishielectric.co.jp/fa/download/search.do?mode=keymanual&q=sh080003)


## 1.動作環境

* OS: Linux
* CPU: ARM/AMD/Intel  

## 2.対応している接続方式
* Ethernet接続


## 3.IO

### Input
PLCのレジスタへの読み取りを定期実行し、16進数のバイナリで構成された電文を取得します。

### Output
電文の内容を元にkanban(RabbitMQ)へデータの投入を行います。

## 4.PLCの読み取り
### 電文フォーマット仕様
読み取りの仕様は下記の通りです。

対応フォーマット：3Eフレーム（固定）
接続先ネットワーク：マルチドロップ局（固定）
読み取り方式：バイト単位の一括読み取り（固定）
自局番号：00（固定）

### デバイス番号
読み取るデバイスのデバイス番号はyamlファイルで設定が可能です。

yamlファイルは`/var/lib/aion/default/config/`へ設置してください。

#### 書き方
```
strContent: デバイス名
iDataSize: データ長
strDevNo: デバイス番号
iReadWrite: IO（IN:0, OUT: 1）
iFlowNo: 実行フロー番号
```

例:
```yaml
settings:
  - strContent: "sample"
    iDataSize: 16
    strDevNo: X9000
    iReadWrite: 0
    iFlowNo: 0
  - strContent: "sample2"
    iDataSize: 16
    strDevNo: X9020
    iReadWrite: 0
    iFlowNo: 0
```


#### 設定手順
```shell
mv nis_settings.yaml.sample nis_settings.yaml

# nis_settings.yamlを書き換え

cp nis_settings.yaml /var/lib/aion/default/config/nis_setting.yaml
```

## 5.関連するマイクロサービス
control-mitsubishi-plc-w-kube