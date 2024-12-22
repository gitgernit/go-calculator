package http

import (
	"encoding/json"
	"fmt"
	"github.com/gitgernit/go-calculator/internal/domain/calculator"
	"net/http"
)

type RequestBody struct {
	Expression string `json:"expression"`
}

func CalculateHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writer.Header().Set("Content-Type", "application/json")

	var reqBody RequestBody
	err := json.NewDecoder(request.Body).Decode(&reqBody)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusUnprocessableEntity)
		err := json.NewEncoder(writer).Encode(map[string]string{"error": "Expression is not valid"})
		if err != nil {
			panic(err)
		}
		return
	}

	interactor := calculator.NewCalculatorInteractor()
	result, err := interactor.Calculate(reqBody.Expression)

	if err != nil {
		writer.WriteHeader(http.StatusUnprocessableEntity)
		err := json.NewEncoder(writer).Encode(map[string]string{"error": "Expression is not valid"})
		if err != nil {
			panic(err)
		}
		return
	}

	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(map[string]interface{}{
		"result": result,
	})
	if err != nil {
		panic(err)
	}
}
