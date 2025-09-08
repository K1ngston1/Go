package main

import "fmt"

func main() {
	p1 := Patient{"Ivan", 18, "Flu ", 38}
	p2 := Patient{"Olga", 13, "Flu", 36}
	p3 := Patient{"Oleg", 18, "Flu", 39}
	p4 := Patient{"Ivan", 30, "Flu", 38}

	stream := Stream[Patient]{}
	stream.Add(p1)
	stream.Add(p2)
	stream.Add(p3)
	stream.Add(p4)

	fmt.Println("All patients:")
	stream.DisplayAll()

	older := stream.Filter(func(p Patient) bool { return p.age > 25 })
	fmt.Println("\nFiltered (age > 25):")
	older.DisplayAll()

	aged := MapStream(stream, func(p Patient) Patient {
		p.age *= 2
		return p
	})
	fmt.Println("\nMapped (age*2):")
	aged.DisplayAll()

	maxPatient := Max(stream, func(p Patient) int { return p.age })
	fmt.Println("\nMax age patient:")
	fmt.Println(maxPatient.Display())

	totalAge := Reduce(stream, 0, func(acc int, p Patient) int { return acc + p.age })
	fmt.Println("\nTotal age:", totalAge)

	distinctPatients := stream.Distinct()
	fmt.Println("\nDistinct patients:")
	distinctPatients.DisplayAll()
}

type Displayable interface {
	Display() string
}

type Stream[T Displayable] struct {
	data []T
}
type Patient struct {
	name    string
	age     int
	history string
	degree  int
}

func (p Patient) Display() string {
	return fmt.Sprintf("Patient: %s, History: %s, Degree: %d", p.name, p.history, p.degree)
}

type Hospital struct {
	name       string
	Address    string
	Department string
}

func (h Hospital) Display() string {
	return fmt.Sprintf("Hospital: %s, Address: %s, Department: %s", h.name, h.Address, h.Department)
}

type Doctor struct {
	name           string
	age            int
	specialization string
	schedule       string
}

func (d Doctor) Display() string {
	return fmt.Sprintf("Doctor: %s, Specialization: %s", d.name, d.specialization)
}

type Department struct {
	Hospital  Hospital
	ward      string
	equipment string
}

func (d Department) Display() string {
	return fmt.Sprintf("Department: %s, Ward: %s , Equipment: %s", d.Hospital.Display(), d.ward, d.equipment)
}

func (s *Stream[T]) Add(item T) {
	s.data = append(s.data, item)
}

func (s Stream[T]) DisplayAll() {
	for _, v := range s.data {
		fmt.Println(v.Display())
	}
}

func (s Stream[T]) Filter(pred func(T) bool) Stream[T] {
	var filtered Stream[T]
	for _, v := range s.data {
		if pred(v) {
			filtered.Add(v)
		}
	}
	return filtered
}

func MapStream[T Displayable, R Displayable](s Stream[T], f func(T) R) Stream[R] {
	var mapped Stream[R]
	for _, v := range s.data {
		mapped.Add(f(v))
	}
	return mapped
}

func Max[T Displayable](s Stream[T], key func(T) int) *T {
	if len(s.data) == 0 {
		return nil
	}
	maxElem := s.data[0]
	maxVal := key(maxElem)
	for _, v := range s.data[1:] {
		if key(v) > maxVal {
			maxElem = v
			maxVal = key(v)
		}
	}
	return &maxElem
}

func Reduce[T Displayable, R any](s Stream[T], init R, f func(R, T) R) R {
	result := init
	for _, v := range s.data {
		result = f(result, v)
	}
	return result
}

func (s Stream[T]) Distinct() Stream[T] {
	seen := make(map[string]bool)
	var distinct Stream[T]
	for _, v := range s.data {
		key := v.Display()
		if !seen[key] {
			seen[key] = true
			distinct.Add(v)
		}
	}
	return distinct
}
