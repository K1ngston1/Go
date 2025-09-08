package main

import "fmt"

func main() {
	schoolGrades := map[string][]int{"Alex": {1, 2, 4}, "Egor": {5, 4, 3}}
	fmt.Println("Оцінки Alex:", schoolGrades["Alex"])

	var position string
	var student string

	for {
		fmt.Println("Введіть позицію (1-5): 1)Створити студента 2)Додати оцінку студента 3)вивести студента з оцінками 4)вивести середню оцінку 5) вийти з програми ")

		fmt.Scan(&position)

		switch position {
		case "1":
			fmt.Print("Введіть ім'я нового студента: ")
			fmt.Scan(&student)
			if _, exists := schoolGrades[student]; exists {
				fmt.Println("Такий студент вже існує.")
			} else {
				schoolGrades[student] = []int{}
				fmt.Println("Студента додано:", student)
			}
		case "2":
			fmt.Print("Введіть ім'я студента для додавання оцінки: ")
			fmt.Scan(&student)
			if _, exists := schoolGrades[student]; !exists {
				fmt.Println("Студент не знайдений.")
				continue
			}
			var grade int
			fmt.Print("Введіть оцінку: ")
			fmt.Scan(&grade)
			schoolGrades[student] = append(schoolGrades[student], grade)
			fmt.Println("Оцінку додано.")
		case "3":
			fmt.Print("Введіть ім'я студента для вивода оцінки")
			fmt.Scan(&student)
			if grades, exists := schoolGrades[student]; exists {
				fmt.Println("Оцінка:", grades, "Студент", student)
			} else {
				fmt.Println("Студент не знайдений.")
			}
		case "4":
			fmt.Print("Введіть ім'я студента для середньої оцінки: ")
			fmt.Scan(&student)
			if grades, exists := schoolGrades[student]; exists {
				if len(grades) == 0 {
					fmt.Println("У студента немає оцінок.")
				} else {
					sum := 0
					for _, g := range grades {
						sum += g
					}
					avg := float64(sum) / float64(len(grades))
					fmt.Printf("Середня оцінка %s: %.2f\n", student, avg)
				}
			} else {
				fmt.Println("Студент не знайдений.")
			}
		case "5":
			fmt.Print("Вихід з  програми")
			return
		default:
			fmt.Println("Невірна опція")
		}
	}
}
