package messages

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MessageSender interface {
	SendMessage(text string, userID int64) error
}

type Model struct {
	tgClient MessageSender
}

func New(tgClient MessageSender) *Model {
	return &Model{
		tgClient: tgClient,
	}
}

const regularAnswer = `
use some of these commands:
/start - to start the convo
/add [your amount] - to add an amount
/sub [your amount] - to substract an amount
/history - to see your finance history
`

const wrongValue = `
try another value
`

const choosePeriod = `
choose one of these options:
/day
/week
/month
/year
/all - to see all history of funds
`

var users = make(map[int64]Info)

type Message struct {
	Text     string
	UserID   int64
	UserName string
}

type Info struct {
	Position string
	Current  int
	History  []Date
}

type Date struct {
	Day    time.Time
	Amount int
}

func provideInfo(msg Message) string {
	if _, ok := users[msg.UserID]; !ok {
		return startConvo(msg.UserID, msg.UserName)
	}

	if users[msg.UserID].Position != "history" {
		return startConvo(msg.UserID, msg.UserName)
	}

	var countDown time.Time

	switch msg.Text {
	case "/day":
		countDown = time.Now().AddDate(0, 0, -1)
	case "/week":
		countDown = time.Now().AddDate(0, 0, -7)
	case "/month":
		countDown = time.Now().AddDate(0, -1, 0)
	case "/year":
		countDown = time.Now().AddDate(-1, 0, 0)
	case "/all":
		countDown = time.Date(2000, 1, 1, 1, 1, 1, 1, time.FixedZone("Astana", 5*60*60))
	}

	var res string
	var total int

	for _, date := range users[msg.UserID].History {
		if date.Day.After(countDown) {
			res += date.Day.Format("01-February-06 3:04PM") + "\t" + strconv.Itoa(date.Amount) + "\n"
		}
		total += date.Amount
	}

	users[msg.UserID] = Info{
		Position: "home",
		Current:  users[msg.UserID].Current,
		History:  users[msg.UserID].History,
	}

	return res + "\n" + formatNumber(total) + "\n" + regularAnswer
}

func formatNumber(n int) string {
	var res string
	s := strconv.Itoa(n)
	temp := ""
	for i := len(s) - 1; i >= 0; i-- {
		temp = string(s[i]) + temp
		if len(temp) == 3 {
			res += " " + temp
			temp = ""
		}
	}
	if temp != "" {
		res = temp + res
	}

	return res + " tenge"
}

func history(msg Message) string {
	if _, ok := users[msg.UserID]; !ok {
		return startConvo(msg.UserID, msg.UserName)
	}

	users[msg.UserID] = Info{
		Position: "history",
		Current:  users[msg.UserID].Current,
		History:  users[msg.UserID].History,
	}

	return choosePeriod
}

func subFunds(msg Message) string {
	if _, ok := users[msg.UserID]; !ok {
		return regularAnswer
	}

	users[msg.UserID] = Info{
		Position: "subFunds",
		Current:  users[msg.UserID].Current,
		History:  users[msg.UserID].History,
	}

	amount, err := strconv.Atoi(regexp.MustCompile(`\d+`).FindString(msg.Text))
	if err != nil {
		return wrongValue
	}

	users[msg.UserID] = Info{
		Position: "subFunds",
		Current:  users[msg.UserID].Current - amount,
		History:  append(users[msg.UserID].History, Date{time.Now(), -1 * amount}),
	}
	return "funds substracted!"
}

func addFunds(msg Message) string {
	if _, ok := users[msg.UserID]; !ok {
		return regularAnswer
	}

	users[msg.UserID] = Info{
		Position: "addFunds",
		Current:  users[msg.UserID].Current,
		History:  users[msg.UserID].History,
	}

	amount, err := strconv.Atoi(regexp.MustCompile(`\d+`).FindString(msg.Text))
	fmt.Println(amount)
	if err != nil {
		return wrongValue
	}

	users[msg.UserID] = Info{
		Position: "addFunds",
		Current:  users[msg.UserID].Current + amount,
		History:  append(users[msg.UserID].History, Date{time.Now(), amount}),
	}
	return "funds added!"
}

func startConvo(userID int64, userName string) string {
	if _, ok := users[userID]; ok {
		return regularAnswer
	}
	users[userID] = Info{
		Position: "startConvo",
	}
	return "hello " + userName
}

func (s *Model) IncomingMessage(msg Message) error {
	switch {
	case msg.Text == "/home":
		return s.tgClient.SendMessage(startConvo(msg.UserID, msg.UserName), msg.UserID)
	case msg.Text == "/start":
		return s.tgClient.SendMessage(startConvo(msg.UserID, msg.UserName), msg.UserID)
	case strings.HasPrefix(msg.Text, "/add"):
		return s.tgClient.SendMessage(addFunds(msg), msg.UserID)
	case strings.HasPrefix(msg.Text, "/sub"):
		return s.tgClient.SendMessage(subFunds(msg), msg.UserID)
	case msg.Text == "/history":
		return s.tgClient.SendMessage(history(msg), msg.UserID)
	case msg.Text == "/day" || msg.Text == "/week" || msg.Text == "/month" || msg.Text == "/year" || msg.Text == "/all":
		return s.tgClient.SendMessage(provideInfo(msg), msg.UserID)
	default:
		return s.tgClient.SendMessage(regularAnswer, msg.UserID)
	}
}
