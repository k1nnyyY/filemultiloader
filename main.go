//--- cd /root/new/filemultiloader/
// go run main.go
 
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)
type Task struct {
	UserID int    `json:"userId"`
	URL    string `json:"url"`
	Method string `json:"method"`
	ID     string `json:"id"`
	Status string `json:"status"`
}

func DownloadFileWithProgressDirect(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	progress := &Progress{
		Total: fileSize,
	}
	go displayProgress(progress)

	_, err = io.Copy(out, io.TeeReader(resp.Body, progress))
	if err != nil {
		return err
	}

	progress.Complete()

	return nil
}
func DownloadFileWithProgressTor(filepath string, url string) error {
	cmd := fmt.Sprintf("curl --proxy socks5h://localhost:9050 -o %s %s", filepath, url)
	//output, err := exec.Command("bash", "-c", cmd).Output()
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
	 return err
	}
   
	//fmt.Println(string(output))
	return nil
   }

// func DownloadFileWithProgressTor(filepath string, tempUrl string) error {
// 	// Создаем Tor SOCKS прокси URL
// 	proxyUrl, err := url.Parse("socks5://127.0.0.1:9050")
// 	if err != nil {
// 	 return err
// 	}
   
// 	// Создаем транспорт с настройкой Tor SOCKS прокси
// 	transport := &http.Transport{
// 	 Proxy: http.ProxyURL(proxyUrl),
// 	}
   
// 	// Создаем клиент с настроенным транспортом
// 	client := &http.Client{
// 	 Transport: transport,
// 	}
   
// 	// Отправляем GET запрос через Tor SOCKS прокси
// 	resp, err := client.Get(tempUrl)
// 	if err != nil {
// 	 return err
// 	}
// 	defer resp.Body.Close()
   
// 	fileSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
// 	if err != nil {
// 	 return err
// 	}
   
// 	out, err := os.Create(filepath)
// 	if err != nil {
// 	 return err
// 	}
// 	defer out.Close()
   
// 	progress := &Progress{
// 	 Total: fileSize,
// 	}
// 	go displayProgress(progress)
   
// 	_, err = io.Copy(out, io.TeeReader(resp.Body, progress))
// 	if err != nil {
// 	 return err
// 	}
   
// 	progress.Complete()
   
// 	return nil
// }

type Progress struct {
	Total       int
	BytesLoaded int
}

func (p *Progress) Write(b []byte) (int, error) {
	n := len(b)
	p.BytesLoaded += n
	return n, nil
}

func (p *Progress) Print() {
	percent := (p.BytesLoaded * 100) / p.Total
	fmt.Printf("\rЗагружено %d%%", percent)
}

func (p *Progress) Complete() {
	p.Print() // Выводим процент загрузки перед сообщением о завершении
	fmt.Println("\nЗагрузка завершена")
}

func displayProgress(p *Progress) {
	for {
		p.Print()
	}
}


func getTasks() ([]Task, error) {
	url := "http://localhost:4001/api/tasks/get_list"
   
	// Отправляем GET запрос
	resp, err := http.Get(url)
	if err != nil {
	 return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()
   
	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	 return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	// Декодируем JSON в структуру
	var tasks []Task
	err = json.Unmarshal(body, &tasks)
	if err != nil {
	 return nil, fmt.Errorf("ошибка при декодировании JSON: %v", err)
	}

	return tasks, nil
   }

func setStatusTask(itemUserID int, itemUrl, itemID, itemStatus string) {
	// cmd := "ls -l"
	cmd := "curl -XPOST http://localhost:4001/api/tasks/set_task -H 'Content-Type: application/json' -d '{\"userId\": " + strconv.Itoa(itemUserID) + ", \"url\": \"" + itemUrl + "\", \"method\": \"direct\", \"id\": \"" + itemID + "\", \"status\": \"" + itemStatus + "\"}'"
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println("Ошибка выполнения команды:", err)
		return
	}
	fmt.Println("Результат выполнения команды:")
	fmt.Println(string(output))
}
func DeleteTask(itemID string) {
	//cmd := "curl -XPOST http://localhost:4001/api/tasks/set_task -H 'Content-Type: application/json' -d '{\"userId\": " + strconv.Itoa(itemUserID) + ", \"url\": \"" + itemUrl + "\", \"method\": \"direct\", \"id\": \"" + itemID + "\", \"status\": \"" + itemStatus + "\"}'"
	cmd := "curl -XPOST http://localhost:4001/api/tasks/rm_task -H 'Content-Type: application/json' -d '{\"id\": \""+itemID+"\"}'"
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println("Ошибка выполнения команды:", err)
		return
	}
	fmt.Println("Результат выполнения команды:")
	fmt.Println(string(output))
}
func Hello(wg *sync.WaitGroup, numbersChan <-chan int, tasks []Task) {
	defer wg.Done()
	for thread_num := range numbersChan {
		if len(tasks)>0{
			fmt.Println("===========================", thread_num)
			for i := 0; i < len(tasks); i++ {
				fmt.Println("------------------")
				itemUrl := tasks[i].URL
				itemID := tasks[i].ID
				itemUserID := tasks[i].UserID
				itemMethod := tasks[i].Method
				itemStatus := tasks[i].Status
				temp_array := strings.Split(itemUrl, "/")

				if itemStatus!="DOWNLOADED"{

					// Получаем последний элемент массива
					fileName := temp_array[len(temp_array)-1]
					fmt.Println("itemUrl: ", itemUrl)
					fmt.Println("itemStatus: ", itemStatus)
					fmt.Println("Качаем файл: ", fileName)

					itemStatus="INPROCESS"
					setStatusTask(itemUserID, itemUrl, itemID, itemStatus)
					// Вывод результатов выполнения команды


					fmt.Println("Качаем файл: ", fileName)

					dirPath := "/root/new/filemultiloader/downloaded/"+strconv.Itoa(itemUserID)+"/"+itemID+"/"

					// Проверяем существование папки
					if _, err := os.Stat(dirPath); os.IsNotExist(err) {
					 // Папка не существует, поэтому создаем ее
					 err := os.MkdirAll(dirPath, 0755)
					 if err != nil {
					  fmt.Println("Ошибка при создании папки:", err)
					  return
					 }
					 fmt.Println("Папка успешно создана!")
					} else {
					 // Папка уже существует
					 fmt.Println("Папка уже существует!")
					}
					// itemMethod = "tor"
					switch {
						case (itemMethod == "direct"):
							err := DownloadFileWithProgressDirect( dirPath + fileName, itemUrl)
							if err != nil {
								fmt.Println("Ошибка при скачивании файла (direct): ", err)
								return
							}
						case (itemMethod == "tor"):
							err := DownloadFileWithProgressTor( dirPath + fileName, itemUrl)
							if err != nil {
								fmt.Println("Ошибка при скачивании файла (tor): ", err)
								return
							}
						// case (itemMethod == "proxy"):
						// 	err := DownloadFileWithProgressProxy( dirPath + fileName, itemUrl)
						// 	if err != nil {
						// 		fmt.Println("Ошибка при скачивании файла (proxy): ", err)
						// 		return
						// 	}
					}
					itemStatus="DOWNLOADED"
					setStatusTask(itemUserID, itemUrl, itemID, itemStatus)
				}
			}
		}

	}
}

func main() {

	// fmt.Println("***********   getTasks   *************")
	tasks, err := getTasks()
	if err != nil {
	 fmt.Println(err)
	 return
	}
    
    atLeastOne := false
	// Текущий список задач
	for _, task := range tasks {
	 fmt.Printf("UserID: %d\nURL: %s\nMethod: %s\nID: %s\nStatus: %s\n\n",
	  task.UserID, task.URL, task.Method, task.ID, task.Status)
	  if task.Status == "CREATED"{
		atLeastOne = true
	  }
	}

	// // Удаляем все задачи
	// for _, task := range tasks {
	// 	DeleteTask(task.ID)
	// 	fmt.Printf("Deleted: ",task.ID)
    // }
	// os.Exit(1)

	
	//=============================================================
	//=============================================================
	//=============================================================
	if atLeastOne == true{
		fmt.Println("***********   tasks   *************")
		fmt.Println(tasks)
		fmt.Println("***********   tasks   *************")

		numThreads := runtime.NumCPU()
		runtime.GOMAXPROCS(numThreads)
		//------------------------------------------------

		// Вычисляем размер каждого подмассива
		arrayLen := len(tasks)
		subArrayLen := arrayLen / numThreads

		// Создаем и заполняем банчи
		subArrays := make([][]Task, numThreads)
		for i := 0; i < numThreads; i++ {
			startIndex := i * subArrayLen
			endIndex := (i + 1) * subArrayLen
			if i == numThreads-1 {
				endIndex = arrayLen
			}
			subArrays[i] = tasks[startIndex:endIndex]
			fmt.Println("startIndex: ", startIndex, "   endIndex: ", endIndex)
		}

		// Выводим результат
		fmt.Println(subArrays)
		// os.Exit(1)
		fmt.Println("************************")

		var wg sync.WaitGroup
		wg.Add(numThreads)

		numbersChan := make(chan int, numThreads) // Буферизованный  канал

		go func() {
			for i := 0; i < numThreads; i++ {
				numbersChan <- i
			}
			close(numbersChan)
		}()

		for i := 0; i < numThreads; i++ {
			go Hello(&wg, numbersChan, subArrays[i])
		}
		wg.Wait()
		fmt.Println("All done!")
	}

}
