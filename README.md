# data-platform-api-division-exconf-rmq-kube
data-platform-api-division-exconf-rmq-kube は、データ連携基盤において、製品部門の存在性チェックを行うためのレポジトリです。

## 動作環境
・ エッジ Kubernetes （推奨）    
・ AION のリソース （推奨)    
・ OS: LinuxOS （必須）    
・ CPU: ARM/AMD/Intel（いずれか必須）  
・ RabbitMQ on Kubernetes  
・ RabbitMQ Client

## 存在確認先テーブル名
以下のsqlファイルに対して、製品部門の存在確認が行われます。

* sap-division-sql-division-data.sql （データ連携基盤 製品部門 - 製品部門データ）

## existence_check.go による存在性確認
Input で取得されたファイルに基づいて、existence_check.go で、database 内の models 内の data_platform_division_data.go から製品部門の対象値の存在性確認が行われます。

```
func ExistenceCheck(ctx context.Context, db *database.Mysql, partnerId string) (bool, error) {
	d, err := models.FindDataPlatformBusinessPartnerGeneralDatum(
		ctx, db, partnerId, models.DataPlatformBusinessPartnerGeneralDatumColumns.BusinessPartner,
	)
	if err != nil {
		return false, xerrors.Errorf("cannot get data: %w", err)
	}
	if d == nil {
		return false, nil
	}
	if d.BusinessPartner != partnerId {
		return false, nil
	}
	return true, nil
}
```

## Input
data-platform-api-division-exconf-rmq-kube は、入力ファイルとして、RabbitMQ からのメッセージを JSON 形式で受け取ります。入力ファイルは、input_files 内の test.json にあります。

```
{
	"connection_key": "request",
	"result": true,
	"redis_key": "abcdefg",
	“runtime_session_id”:
	”boi9ar543dg91ipdnspi099u231280ab0v8af0ew”,
	"filepath": "/var/lib/aion/Data/rededge_sdc/abcdef.json",
	"service_label": "ORDERS",
	"Orders": {
		"Division": "01"
	},
	"api_schema": "DPFMOrdersCreates",
	"accepter": ["All"],
	"order_id": null,
	"deleted": false
}
```

## RabbitMQ からのメッセージ受信による イベントドリヴン の ランタイム実行
data-platform-api-division-exconf-rmq-kube は、RabbitMQ からのメッセージを受け取ると、イベントドリヴンでランタイムを実行します。  
AION の仕様では、Kubernetes 上 の 当該マイクロサービスPod は 立ち上がったまま待機状態で当該メッセージを受け取り、（コンテナ起動などの段取時間をカットして）即座にランタイムを実行します。　 


## RabbitMQ の マスタサーバ環境
data-platform-api-division-exconf-rmq-kube が利用する RabbitMQ のマスタサーバ環境は、rabbitmq-on-kubernetes です。  
当該マスタサーバ環境は、同じエッジコンピューティングデバイスに配置されても、別の物理(仮想)サーバ内に配置されても、どちらでも構いません。

## RabbitMQ の Golang Runtime ライブラリ
data-platform-api-division-exconf-rmq-kube は、RabbitMQ の Golang Runtime ライブラリ として、rabbitmq-golang-clientを利用しています。

## デプロイ・稼働
data-platform-api-division-exconf-rmq-kube の デプロイ・稼働 を行うためには、aion-service-definitions の services.yml に、本レポジトリの services.yml を設定する必要があります。

kubectl apply - f 等で Deployment作成後、以下のコマンドで Pod が正しく生成されていることを確認してください。

```
$ kubectl get pods
```


## Output
data-platform-api-division-exconf-rmq-kube では、[golang-logging-library](https://github.com/latonaio/golang-logging-library) により、Output として、RabbitMQ へのメッセージを JSON 形式で出力します。ビジネスパートナの対象値が存在する場合 true、存在しない場合 false、を返します。"cursor" ～ "time"は、golang-logging-library による 定型フォーマットの出力結果です。

```
{
    "cursor": "/go/src/github.com/latonaio/data-platform-api-division-exconf/main.go#L56",
    "function": "('DPFM_API_ORDERS_SRV', 'creates', 'A_Item', 'Division', 'DPFM_API_DIVISION_SRV', 'exconf');",
    "level": "INFO",
    "service_label": "ORDERS",
    "Product": {
        "Division": "01",
        "ExistenceConf": true
    },
    "runtime_session_id": "boi9ar543dg91ipdnspi099u231280ab0v8af0ew",
    "time": "2022-09-29T07:04:23Z"
}
```