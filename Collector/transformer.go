package Collector

import (
	"BHLayer2Node/paradigm"
	"encoding/json"
	"github.com/go-gota/gota/dataframe"
)

// OutputTransformer 根据输出类型，将传入的数据还原
type OutputTransformer struct {
	outputType paradigm.ModelOutputType
}

// Transform 将[]byte根据输出类型转换
func (t *OutputTransformer) Transform(data []byte) (interface{}, error) {
	switch t.outputType {
	case paradigm.DATAFRAME:
		// 如果是dataframe，那么将data解析出json,然后将json转换为dataframe
		// dataframe在python中得到的json, 如是是多条df会用嵌套字典的方法叠加，所以要额外处理
		jsonStr, err := t.parseJson(data)
		if err != nil {
			return nil, err
		}
		return t.TransformDataFrame(jsonStr)
		// todo 剩下的类型适配
	default:
		panic("Unknown Output Type!!!")
	}
}
func (t *OutputTransformer) TransformDataFrame(rawData map[string]map[string]interface{}) (dataframe.DataFrame, error) {

	cols := make([]string, 0)
	for _, values := range rawData {
		for col, _ := range values {
			cols = append(cols, col)
		}
		break
		// 得到col
	}

	// 创建一个空的切片来存储每行数据
	var result []map[string]interface{}

	// 遍历每一行数据，构造每一行的 map
	for _, col := range cols {
		// 创建一个新的 map 来存储当前行的数据
		row := make(map[string]interface{})

		// 填充每一列的数据到当前行
		for key, values := range rawData {
			// 将列的索引（i）作为字符串来获取对应的数据
			row[key] = values[col]
		}

		// 将当前行数据添加到结果中
		result = append(result, row)
	}
	//fmt.Println(rawData)
	//fmt.Println(result)
	// 转换为 DataFrame
	df := dataframe.LoadMaps(result)
	//fmt.Println(df)
	return df, nil
}

// parseJson 解析byte->json
func (t *OutputTransformer) parseJson(jsonStr []byte) (map[string]map[string]interface{}, error) {
	var rawData map[string]map[string]interface{}
	err := json.Unmarshal(jsonStr, &rawData)
	if err != nil {
		return nil, err
	}
	return rawData, nil
}

//func (t *OutputTransformer) BatchTransform(datas [][]byte) ([]interface{}, error) {
//	jsonStrs := make([]map[string]map[string]interface{}, 0)
//	for _, data := range datas {
//		jsonStr, err := t.parseJson(data)
//		if err != nil {
//			panic("parse json error!!!")
//		}
//		jsonStrs = append(jsonStrs, jsonStr)
//	}
//	for _, jsonstr := range jsonStrs {
//		t.Transform(jsonstr)
//	}
//}
