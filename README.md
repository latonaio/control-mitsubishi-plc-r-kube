# control-mitsubishi-plc-w-kube
三菱電機製のPLCのレジスタに登録されたメッセージを読み込むマイクロサービスです。

メッセージの送受信方法およびフォーマットはMCプロトコルに準じています。

### 動作環境

* OS: Linux
* CPU: Intel64/AMD64/ARM64

### 対応している接続方式
* Ethernet接続 （シリアル接続は非対応）

## MCプロトコル
三菱電機製レジスタに採用されている、三菱電機独自のプロトコルです。

16進数のバイナリで構成された電文を送受信し、レジスタに対して操作を行うメッセージングプロトコルです。

Ethernetおよびシリアル接続に対応しています。

[MCプロトコルのマニュアル（三菱電機のHPに遷移します）](https://www.mitsubishielectric.co.jp/fa/download/search.do?mode=keymanual&q=sh080003)


## Input
PLCのレジスタへの読み取りを定期実行し、16進数のバイナリで構成された電文を取得します。

## Output
電文の内容を元にkanbanへデータの投入を行います。

##関連するマイクロサービス
control-mitsubishi-plc-w-kube