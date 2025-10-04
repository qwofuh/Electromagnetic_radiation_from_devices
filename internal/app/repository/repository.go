package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Order struct { // вот наша новая структура
	ID               int    // поля структур, которые передаются в шаблон
	Title            string // ОБЯЗАТЕЛЬНО должны быть написаны с заглавной буквы (то есть публичными)
	Image            string
	AvgMinPower      float64 // Типовая мощность
	AvgMaxPower      int
	SafeRange        string // Безопасное расстояние
	RadiationType    string // Тип излучения
	RadiationSource  string // Источник излучения
	MaxRadiationZone string // Зона максимального излучения
}

func (r *Repository) GetOrders() ([]Order, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	orders := []Order{ // массив элементов из наших структур
		{
			ID:               1,
			Title:            "Холодильник",
			Image:            "http://127.0.0.1:9000/devices/FridgeMain.png",
			AvgMinPower:      100,
			AvgMaxPower:      300,
			SafeRange:        "0.5-1 м",
			RadiationType:    "низкочастотное электромагнитное поле",
			RadiationSource:  "компрессор, блок управления",
			MaxRadiationZone: "задняя стенка с компрессором",
		},
		{
			ID:               2,
			Title:            "СВЧ-печь",
			Image:            "http://127.0.0.1:9000/devices/SVCH.jpeg",
			AvgMinPower:      800,
			AvgMaxPower:      1200,
			SafeRange:        "1-2 м",
			RadiationType:    "СВЧ-излучение",
			RadiationSource:  "магнетрон",
			MaxRadiationZone: "вокруг дверцы",
		},
		{
			ID:               3,
			Title:            "WIFI-роутер",
			Image:            "http://127.0.0.1:9000/devices/wifi.avif",
			AvgMinPower:      0.1,
			AvgMaxPower:      1,
			SafeRange:        "2-5 м",
			RadiationType:    "радиочастотное",
			RadiationSource:  "антенны, передатчик",
			MaxRadiationZone: "вокруг антенн",
		},
		{
			ID:               4,
			Title:            "Телевизор",
			Image:            "http://127.0.0.1:9000/devices/TV.jpg",
			AvgMinPower:      50,
			AvgMaxPower:      200,
			SafeRange:        "2-3 м",
			RadiationType:    "низкочастотное магнитное поле",
			RadiationSource:  "блок питания, задняя панель",
			MaxRadiationZone: "задняя часть устройства",
		},
		{
			ID:               5,
			Title:            "Стиральная машина",
			Image:            "http://127.0.0.1:9000/devices/CleaningMachine.png",
			AvgMinPower:      1500,
			AvgMaxPower:      2500,
			SafeRange:        "1-2 м",
			RadiationType:    "низкочастотное магнитное поле",
			RadiationSource:  "электродвигатель, ТЭН",
			MaxRadiationZone: "близко к двигателю и блоку управления",
		},
		{
			ID:               6,
			Title:            "Компьютер",
			Image:            "http://127.0.0.1:9000/devices/PK.jpeg",
			AvgMinPower:      300,
			AvgMaxPower:      800,
			SafeRange:        "0.5-1 м",
			RadiationType:    "низкочастотное магнитное поле",
			RadiationSource:  "блок питания",
			MaxRadiationZone: "задняя часть системнного блока",
		},
		{
			ID:               7,
			Title:            "Индукционная плита",
			Image:            "http://127.0.0.1:9000/devices/induct.jpg",
			AvgMinPower:      2000,
			AvgMaxPower:      3500,
			SafeRange:        "0.5-1 м",
			RadiationType:    "низкочастотное магнитное поле",
			RadiationSource:  "медные катушки под стеклокерамической поверзностью",
			MaxRadiationZone: "непосредственно над включенной конфоркой",
		},
		{
			ID:               8,
			Title:            "Посудомоечная машина",
			Image:            "http://127.0.0.1:9000/devices/posudomoyka.jpeg",
			AvgMinPower:      1800,
			AvgMaxPower:      2200,
			SafeRange:        "1-1.5 м",
			RadiationType:    "низкочастотное магнитное поле",
			RadiationSource:  "блок управления, нагревательный элемент, двигатель",
			MaxRadiationZone: "панель управления",
		},
	}
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(orders) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return orders, nil
}

func (r *Repository) GetOrder(id int) (Order, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	orders, err := r.GetOrders()
	if err != nil {
		return Order{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, order := range orders {
		if order.ID == id {
			return order, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Order{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetOrdersByTitle(title string) ([]Order, error) {
	orders, err := r.GetOrders()
	if err != nil {
		return []Order{}, err
	}

	var result []Order
	for _, order := range orders {
		if strings.Contains(strings.ToLower(order.Title), strings.ToLower(title)) {
			result = append(result, order)
		}
	}

	return result, nil
}
