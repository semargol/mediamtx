package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
)

func main() {
	// Значение sprop-parameter-sets
	sprop := "Z01AH+iAKALdgLUBAQFAAAADAEAAAAyDxgxEgA==,aOvvIA=="

	// Получение SPS и PPS в виде байтовых массивов
	sps, pps, err := getSPSPPS(sprop)
	if err != nil {
		log.Fatalf("Error extracting SPS/PPS: %v", err)
	}

	// Извлечение profile-level-id из SPS
	profileLevelID := strings.ToUpper(hex.EncodeToString(sps[1:4]))

	// Кодирование в base64
	spsBase64 := base64.StdEncoding.EncodeToString(sps)
	ppsBase64 := base64.StdEncoding.EncodeToString(pps)

	// Объединение строк
	spropParameterSets := fmt.Sprintf("%s,%s; profile-level-id=%s", spsBase64, ppsBase64, profileLevelID)

	fmt.Println("SPS:", sps)
	fmt.Println("PPS:", pps)
	fmt.Println("sprop-parameter-sets:", spropParameterSets)
}

// Функция для получения SPS и PPS из строки sprop-parameter-sets
func getSPSPPS(sprop string) ([]byte, []byte, error) {
	// Разделяем строку по запятой
	parts := strings.Split(sprop, ",")
	if len(parts) < 2 {
		return nil, nil, fmt.Errorf("invalid sprop-parameter-sets (%v)", sprop)
	}

	// Декодируем первое значение (SPS)
	sps, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid base64 string for SPS: %v", err)
	}

	// Декодируем второе значение (PPS)
	pps, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid base64 string for PPS: %v", err)
	}

	return sps, pps, nil
}
