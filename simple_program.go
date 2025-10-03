package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("🎯 Добро пожаловать в простую программу!")

	// Генерация случайного числа
	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(100) + 1

	fmt.Printf("🎲 Случайное число: %d\n", number)

	// Простая логика
	if number > 50 {
		fmt.Println("✅ Больше 50 - отлично!")
	} else {
		fmt.Println("❌ Меньше или равно 50 - попробуйте еще раз!")
	}

	// Массив с разными сообщениями
	messages := []string{
		"🚀 Космос ждет!",
		"🌟 Звезды сияют!",
		"🌙 Луна улыбается!",
		"☀️ Солнце светит!",
		"🌈 Радуга появилась!",
	}

	// Выбор случайного сообщения
	randomMessage := messages[rand.Intn(len(messages))]
	fmt.Printf("💫 %s\n", randomMessage)

	fmt.Println("🎉 Программа завершена!")
}
