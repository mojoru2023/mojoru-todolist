package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Todo struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Content string `json:"content"`
}

// 初始化数据库
func initDB() (*gorm.DB, error) {
	dbDir := "DB"
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		os.Mkdir(dbDir, os.ModePerm)
	}

	dbPath := filepath.Join(dbDir, "todo.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Todo{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// 查询待办事项
func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
	if err != nil {
		http.Error(w, "数据库错误", http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	var todos []Todo
	result := db.Find(&todos)
	if result.Error != nil {
		http.Error(w, "查询数据失败", http.StatusInternalServerError)
		log.Fatal(result.Error)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// 保存待办事项到数据库
func saveTodosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var todos []Todo
		err := json.NewDecoder(r.Body).Decode(&todos)
		if err != nil {
			http.Error(w, "无法解析请求体", http.StatusBadRequest)
			return
		}

		db, err := initDB()
		if err != nil {
			http.Error(w, "数据库错误", http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		// 清除原有待办事项
		if err := db.Exec("DELETE FROM todos").Error; err != nil {
			http.Error(w, "删除数据失败", http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		// 插入新的待办事项
		for _, todo := range todos {
			result := db.Create(&todo)
			if result.Error != nil {
				http.Error(w, "插入数据失败", http.StatusInternalServerError)
				log.Fatal(result.Error)
				return
			}
		}

		response := map[string]interface{}{
			"message": "待办事项已保存",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	} else {
		http.Error(w, "不支持的方法", http.StatusMethodNotAllowed)
	}
}

// 显示 HTML 模板
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		http.Error(w, "无法加载模板", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/save", saveTodosHandler) // 保存待办事项的路由
	http.HandleFunc("/todos", getTodosHandler) // 获取待办事项的路由

	fmt.Println("服务器正在运行，访问地址: http://localhost:8004")
	log.Fatal(http.ListenAndServe(":8004", nil))
}
