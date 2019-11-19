Nanda-Kanda-DB
====

csv なんかを雑に kvs ぽく扱いたいときになんだかんだで使えるかもしれない go ライブラリ。

すべてのレコードをメモリに置いても平気なときだけ使える。

## Example

```
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/thamaji/nkdb"
	"github.com/thamaji/nkdb/storage/cache"
	"github.com/thamaji/nkdb/storage/csv"
)

func main() {
	// データ保存先
	// nkdb.Storage を実装した何か
	storage := cache.New(csv.New("data.csv"), 15*time.Minute)

	// DB作成
	// 第１引数はキー列（0 はじまり）
	db := nkdb.New(0, storage)

	// レコードを登録
	if err := db.Set("a", []string{"a", "1"}); err != nil {
		log.Fatalln(err)
	}

	// レコードを取得
	record, err := db.Get("a")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(record)

	// レコードを削除
	if err := db.Delete("a"); err != nil {
		log.Fatalln(err)
	}

	// 存在しないレコード
	if _, err := db.Get("a"); err != nil {
		if !nkdb.IsNotExist(err) {
			log.Fatalln(err)
		}

		fmt.Println("a dose not exist")
	}

	// 書き込みトランザクション
	err = db.Update(func(tx *nkdb.Tx) error {
		tx.Set("a", []string{"a", "1"})
		tx.Set("b", []string{"b", "2"})
		tx.Set("c", []string{"c", "3"})

		keys, err := tx.Keys()
		if err != nil {
			return err
		}

		for _, key := range keys {
			record, err := tx.Get(key)
			if err != nil {
				return err
			}

			record[1] += "000"

			if err := tx.Set(key, record); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	// 読み込みトランザクション
	err = db.View(func(tx *nkdb.Tx) error {
		keys, err := tx.Keys()
		if err != nil {
			return err
		}

		for _, key := range keys {
			record, err := tx.Get(key)
			if err != nil {
				return err
			}

			fmt.Println(record)
		}

		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
}
```
