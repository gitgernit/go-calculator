package calc

import (
	"encoding/json"
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
	err = json.NewEncoder(writer).Encode(map[string]float64{
		"result": result,
	})
	if err != nil {
		panic(err)
	}
}
