package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

//Estrutura principal da filas
type Queues struct {
	XMLName xml.Name `xml:"queues"`
	Queues  []Queue  `xml:"queue"`
}

//Estrutura com detalhes da fila
type Queue struct {
	XMLName xml.Name `xml:"queue"`
	Name    string   `xml:"name,attr"`
	Stats   Stats    `xml:"stats"`
}

//Estrutura com detalhes a nível de estado da fila
type Stats struct {
	XMLName       xml.Name `xml:"stats"`
	Size          string   `xml:"size,attr"`
	ConsumerCount string   `xml:"consumerCount,attr"`
	EnqueueCount  string   `xml:"enqueueCount,attr"`
	DequeueCount  string   `xml:"dequeueCount,attr"`
}

//Estrutura completada da fila para json
type Fila struct {
	Name          string `json:"name,omitempty"`
	Size          string `json:"size"`
	ConsumerCount string `json:"consumer"`
	EnqueueCount  string `json:"enqueue"`
	DequeueCount  string `json:"dequeue"`
}

//Variabel global que vai receber as informações
var brokerinfo []Fila

//Criação do server http, com 3 regras de ep
func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", redirect).Methods("GET")
	router.HandleFunc("/healthcheck", getWorking).Methods("GET")
	router.HandleFunc("/brokerinfo/", brokerInfo).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", router))

}

//Caso o acesso seja realizar no "/" é realizado um redirect para "/healthcheck"
func redirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/healthcheck", http.StatusFound)
}

//O EP "/healthcheck" retorna "WORKING"
func getWorking(w http.ResponseWriter, r *http.Request) {

	io.WriteString(w, "WORKING")
}

// Recebe as 'informações da função infoBroker() para 'montar o json' no ep "/brokerinfo/" '
func brokerInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RemoteAddr, r.Method, r.Body, r.RequestURI)
	json.NewEncoder(w).Encode(infoBroker())

}

//Função para ir no endpoint fornecido pelo barramento(.jsp) e realizar o parser das informações
func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read body: %v", err)
	}

	return data, nil
}

//Nesta função vamos 'estrutura no json', a opção comentada é caso deseje realizar o procedimento através da leitura de um arquivo, caso sim basta comentar da linha 101 a 109 que recebe os parametros da função getXML
func infoBroker() []Fila {
	/*	xmlFile, err := os.Open("barramento.xml")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Successfully Opened barramento.xml")
		byteValue, _ := ioutil.ReadAll(xmlFile)
		var queue Queues
		xml.Unmarshal(byteValue, &queue)
	*/

	xmlBytes, err := getXML("http://url_barramento/admin/xml/queues.jsp")
	if err != nil {
		log.Printf("Failed to get XML: %v", err)
	}

	var queue Queues
	xml.Unmarshal(xmlBytes, &queue)

	for i := 0; i < len(queue.Queues); i++ {

		newStruct := &Fila{
			Name:          queue.Queues[i].Name,
			Size:          queue.Queues[i].Stats.Size,
			ConsumerCount: queue.Queues[i].Stats.ConsumerCount,
			EnqueueCount:  queue.Queues[i].Stats.EnqueueCount,
			DequeueCount:  queue.Queues[i].Stats.DequeueCount,
		}

		brokerinfo = append(brokerinfo, *newStruct)
	}
	return brokerinfo
}
