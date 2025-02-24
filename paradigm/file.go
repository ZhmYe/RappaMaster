package paradigm

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/go-gota/gota/dataframe"
)

// DataFrameToCSV 将 Gota 的 DataFrame 转换为 CSV 格式的字节流
func DataFrameToCSV(df dataframe.DataFrame) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	// 写入列名
	columnNames := df.Names()
	writer.Write(columnNames)
	fmt.Println(df)
	// 写入每一行数据
	for i := 0; i < df.Nrow(); i++ {
		row := df.Subset([]int{i}).Records()[1] // 取出一行
		writer.Write(row)
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func DataToFile(data interface{}) ([]byte, error) {
	switch data.(type) {
	case dataframe.DataFrame:
		//Log("DEBUG", "Transform data to dataframe")
		//fmt.Println(data)
		return DataFrameToCSV(data.(dataframe.DataFrame))
	default:
		e := Error(ValueError, "can not convert data to file")
		return []byte{}, fmt.Errorf(e.Error())
	}
}
