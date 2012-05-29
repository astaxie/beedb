package beedb

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Model struct {
	Db		*sql.DB
	TableName	string
	LimitStr	int
	OffsetStr	int
	WhereStr	string
	ParamStr	[]interface{}
	OrderStr	string
	ColumnStr	string
	PrimaryKey	string
	JoinStr		string
	GroupByStr	string
	HavingStr	string
}

/**
 * Add New sql.DB in the future i will add ConnectionPool.Get() 
 */
func New(db *sql.DB) (m Model) {
	m = Model{Db: db, ColumnStr: "*", PrimaryKey: "Id"}
	return
}

func (orm *Model) SetTable(tbname string) *Model {
	orm.TableName = tbname
	return orm
}

func (orm *Model) SetPK(pk string) *Model {
	orm.PrimaryKey = pk
	return orm
}

func (orm *Model) Where(querystring interface{}, args ...interface{}) *Model {
	switch querystring := querystring.(type) {
	case string:
		orm.WhereStr = querystring
	case int:
		orm.WhereStr = fmt.Sprintf("`%v` = ?", orm.PrimaryKey)
		args = append(args, querystring)
	}
	orm.ParamStr = args
	return orm
}

func (orm *Model) Limit(start int, size ...int) *Model {
	orm.LimitStr = start
	if len(size) > 0 {
		orm.OffsetStr = size[0]
	}
	return orm
}

func (orm *Model) Offset(offset int) *Model {
	orm.OffsetStr = offset
	return orm
}

func (orm *Model) OrderBy(order string) *Model {
	orm.OrderStr = order
	return orm
}

func (orm *Model) Select(colums string) *Model {
	orm.ColumnStr = colums
	return orm
}

func (orm *Model) SacnPK(output interface{}) *Model {
	if reflect.TypeOf(reflect.Indirect(reflect.ValueOf(output)).Interface()).Kind() == reflect.Slice {
		sliceValue := reflect.Indirect(reflect.ValueOf(output))
		sliceElementType := sliceValue.Type().Elem()
		for i := 0; i < sliceElementType.NumField(); i++ {
			bb := reflect.ValueOf(sliceElementType.Field(i).Tag)
			if bb.String() == "PK" {
				orm.PrimaryKey = sliceElementType.Field(i).Name
			}
		}
	} else {
		tt := reflect.TypeOf(reflect.Indirect(reflect.ValueOf(output)).Interface())
		for i := 0; i < tt.NumField(); i++ {
			bb := reflect.ValueOf(tt.Field(i).Tag)
			if bb.String() == "PK" {
				orm.PrimaryKey = tt.Field(i).Name
			}
		}
	}
	return orm

}

//The join_operator should be one of INNER, LEFT OUTER, CROSS etc - this will be prepended to JOIN
func (orm *Model) Join(join_operator, tablename, condition string) *Model {
	orm.JoinStr = fmt.Sprintf("%v JOIN %v ON %v", join_operator, tablename, condition)
	return orm
}

func (orm *Model) GroupBy(keys string) *Model {
	orm.GroupByStr = fmt.Sprintf("GROUP BY %v", keys)
	return orm
}

func (orm *Model) Having(conditions string) *Model {
	orm.HavingStr = fmt.Sprintf("HAVING %v", conditions)
	return orm
}

func (orm *Model) Find(output interface{}) error {
	orm.SacnPK(output)
	var keys []string
	results, _ := scanStructIntoMap(output)
	orm.TableName = snakeCasedName(StructName(output))
	for key, _ := range results {
		keys = append(keys, key)
	}
	orm.ColumnStr = strings.Join(keys, ", ")
	orm.Limit(1)
	resultsSlice, err := orm.FindMap()
	if err != nil {
		return err
	}
	if len(resultsSlice) == 0 {
		return nil
	} else if len(resultsSlice) == 1 {
		results := resultsSlice[0]
		scanMapIntoStruct(output, results)
	} else {
		return errors.New("More Then One Records")
	}
	return nil
}

func (orm *Model) FindAll(rowsSlicePtr interface{}) error {
	orm.SacnPK(rowsSlicePtr)
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return errors.New("needs a pointer to a slice")
	}

	sliceElementType := sliceValue.Type().Elem()
	st := reflect.New(sliceElementType)
	var keys []string
	results, _ := scanStructIntoMap(st.Interface())
	orm.TableName = getTableName(rowsSlicePtr)
	for key, _ := range results {
		keys = append(keys, key)
	}
	orm.ColumnStr = strings.Join(keys, ", ")

	resultsSlice, err := orm.FindMap()
	if err != nil {
		return err
	}

	for _, results := range resultsSlice {
		newValue := reflect.New(sliceElementType)
		scanMapIntoStruct(newValue.Interface(), results)
		sliceValue.Set(reflect.Append(sliceValue, reflect.Indirect(reflect.ValueOf(newValue.Interface()))))
	}
	return nil
}

func (orm *Model) FindMap() (resultsSlice []map[string][]byte, err error) {
	sqls := orm.generateSql()
	fmt.Println(sqls)
	s, err := orm.Db.Prepare(sqls)
	if err != nil {
		return nil, err
	}
	res, err := s.Query(orm.ParamStr...)
	if err != nil {
		panic(err)
	}
	fields, err := res.Columns()
	if err != nil {
		panic(err)
	}
	for res.Next() {
		result := make(map[string][]byte)
		var scanResultContainers []interface{}
		for i := 0; i < len(fields); i++ {
			var scanResultContainer interface{}
			scanResultContainers = append(scanResultContainers, &scanResultContainer)
		}
		if err := res.Scan(scanResultContainers...); err != nil {
			panic(err)
		}
		for ii, key := range fields {
			rawValue := reflect.Indirect(reflect.ValueOf(scanResultContainers[ii]))
			//if row is null then return nil
			if rawValue.Interface() == nil {
				result[key] = []byte("")
				continue
			}
			aa := reflect.TypeOf(rawValue.Interface())
			vv := reflect.ValueOf(rawValue.Interface())
			var str string
			switch aa.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				str = strconv.FormatInt(vv.Int(), 10)
				result[key] = []byte(str)
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				str = strconv.FormatUint(vv.Uint(), 10)
				result[key] = []byte(str)
			case reflect.Float32, reflect.Float64:
				str = strconv.FormatFloat(vv.Float(), 'f', -1, 64)
				result[key] = []byte(str)
			case reflect.Slice:
				if aa.Elem().Kind() == reflect.Uint8 {
					result[key] = rawValue.Interface().([]byte)
					break
				}
			//时间类型	
			case reflect.Struct:
				str = rawValue.Interface().(time.Time).Format("2006-01-02 15:04:05.000 -0700")
				result[key] = []byte(str)
			}

		}
		resultsSlice = append(resultsSlice, result)
	}
	return resultsSlice, nil
}

func (orm *Model) generateSql() string {
	a := fmt.Sprintf("SELECT %v FROM %v", orm.ColumnStr, orm.TableName)
	if orm.JoinStr != "" {
		a = fmt.Sprintf("%v %v", a, orm.JoinStr)
	}
	if orm.WhereStr != "" {
		a = fmt.Sprintf("%v WHERE %v", a, orm.WhereStr)
	}
	if orm.GroupByStr != "" {
		a = fmt.Sprintf("%v %v", a, orm.GroupByStr)
	}
	if orm.HavingStr != "" {
		a = fmt.Sprintf("%v %v", a, orm.HavingStr)
	}
	if orm.OrderStr != "" {
		a = fmt.Sprintf("%v ORDER BY %v", a, orm.OrderStr)
	}
	if orm.OffsetStr > 0 {
		a = fmt.Sprintf("%v LIMIT %v, %v", a, orm.OffsetStr, orm.LimitStr)
	} else if orm.LimitStr > 0 {
		a = fmt.Sprintf("%v LIMIT %v", a, orm.LimitStr)
	}
	return a
}

//Execute sql
func (orm *Model) Execute(finalQueryString string, args ...interface{}) (sql.Result, error) {
	rs, err := orm.Db.Prepare(finalQueryString)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	res, err := rs.Exec(args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//if the struct has PrimaryKey == 0 insert else update
func (orm *Model) Save(output interface{}) interface{} {
	orm.SacnPK(output)
	results, _ := scanStructIntoMap(output)
	orm.TableName = snakeCasedName(StructName(output))
	id := results[strings.ToLower(orm.PrimaryKey)]
	delete(results, strings.ToLower(orm.PrimaryKey))
	if reflect.ValueOf(id).Int() == 0 {
		id, err := orm.Insert(results)
		if err != nil {
			return nil
		}
		structPtr := reflect.ValueOf(output)
		structVal := structPtr.Elem()
		structField := structVal.FieldByName(orm.PrimaryKey)
		var v interface{}
		x, err := strconv.Atoi(strconv.FormatInt(id, 10))
		if err != nil {
			return err
		}
		v = x
		structField.Set(reflect.ValueOf(v))
		return nil
	} else {
		condition := fmt.Sprintf("`%v`=?", orm.PrimaryKey)
		orm.Where(condition, id)
		_, err := orm.Update(results)
		if err != nil {
			return err
		}
	}
	return nil
}

//inert one info
func (orm *Model) Insert(properties map[string]interface{}) (int64, error) {
	var keys []string
	var placeholders []string
	var args []interface{}

	for key, val := range properties {
		keys = append(keys, key)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}
	statement := fmt.Sprintf("INSERT INTO `%v` (`%v`) VALUES (%v)",
		orm.TableName,
		strings.Join(keys, "`, `"),
		strings.Join(placeholders, ", "))
	res, err := orm.Execute(statement, args...)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return -1, err
	}

	return id, nil
}

//insert batch info
func (orm *Model) InsertBatch(rows []map[string]interface{}) ([]int64, error) {
	var ids []int64
	if len(rows) <= 0 {
		return ids, nil
	}
	for i := 0; i < len(rows); i++ {
		id, _ := orm.Insert(rows[i])
		ids = append(ids, id)
	}
	return ids, nil
}

// update info
func (orm *Model) Update(properties map[string]interface{}) (int64, error) {
	var updates []string
	var args []interface{}

	for key, val := range properties {
		updates = append(updates, fmt.Sprintf("`%v` = ?", key))
		args = append(args, val)
	}
	args = append(args, orm.ParamStr...)
	var condition string
	if orm.WhereStr != "" {
		condition = fmt.Sprintf("WHERE %v", orm.WhereStr)
	} else {
		condition = ""
	}
	statement := fmt.Sprintf("UPDATE `%v` SET %v %v",
		orm.TableName,
		strings.Join(updates, ", "),
		condition)
	res, err := orm.Execute(statement, args...)
	if err != nil {
		return -1, err
	}
	id, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return id, nil
}

func (orm *Model) Delete(output interface{}) (int64, error) {
	orm.SacnPK(output)
	results, _ := scanStructIntoMap(output)
	orm.TableName = snakeCasedName(StructName(output))
	id := results[strings.ToLower(orm.PrimaryKey)]
	condition := fmt.Sprintf("`%v`='%v'", orm.PrimaryKey, id)
	statement := fmt.Sprintf("DELETE FROM `%v` WHERE %v",
		orm.TableName,
		condition)
	res, err := orm.Execute(statement)
	if err != nil {
		return -1, err
	}
	Affectid, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return Affectid, nil
}

func (orm *Model) DeleteAll(rowsSlicePtr interface{}) (int64, error) {
	orm.SacnPK(rowsSlicePtr)
	orm.TableName = getTableName(rowsSlicePtr)
	var ids []string
	val := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if val.Len() == 0 {
		return 0, nil
	}
	for i := 0; i < val.Len(); i++ {
		results, _ := scanStructIntoMap(val.Index(i).Interface())
		id := results[strings.ToLower(orm.PrimaryKey)]
		switch id.(type) {
		case string:
			ids = append(ids, id.(string))
		case int, int64, int32:
			str := strconv.Itoa(id.(int))
			ids = append(ids, str)
		}
	}
	condition := fmt.Sprintf("`%v` in ('%v')", orm.PrimaryKey, strings.Join(ids, "','"))
	statement := fmt.Sprintf("DELETE FROM `%v` WHERE %v",
		orm.TableName,
		condition)
	fmt.Println(statement)
	res, err := orm.Execute(statement)
	if err != nil {
		return -1, err
	}
	Affectid, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return Affectid, nil
}

func (orm *Model) DelectRow() (int64, error) {
	var condition string
	if orm.WhereStr != "" {
		condition = fmt.Sprintf("WHERE %v", orm.WhereStr)
	} else {
		condition = ""
	}
	statement := fmt.Sprintf("DELETE FROM `%v` %v",
		orm.TableName,
		condition)
	fmt.Println(statement)
	res, err := orm.Execute(statement, orm.ParamStr...)
	if err != nil {
		return -1, err
	}
	Affectid, err := res.RowsAffected()

	if err != nil {
		return -1, err
	}
	return Affectid, nil
}
