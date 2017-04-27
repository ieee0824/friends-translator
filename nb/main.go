package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jbrukh/bayesian"
	mecab "github.com/yukihir0/mecab-go"
)

// classの定義
const (
	P bayesian.Class = "Posi"
	N bayesian.Class = "Nega"
	I bayesian.Class = "Illegal"
)

const fileBase = "nb.dat"

var input = flag.String("i", "", "")

func init() {
	flag.Parse()
}

type classifier struct {
	PN *bayesian.Classifier
	NI *bayesian.Classifier
	IP *bayesian.Classifier
}

func newClassifier() *classifier {
	return &classifier{
		bayesian.NewClassifier(P, N),
		bayesian.NewClassifier(N, I),
		bayesian.NewClassifier(I, P),
	}
}

func load() *classifier {
	PN, err := bayesian.NewClassifierFromFile("pn_" + fileBase)
	if err != nil {
		return nil
	}
	NI, err := bayesian.NewClassifierFromFile("ni_" + fileBase)
	if err != nil {
		return nil
	}
	IP, err := bayesian.NewClassifierFromFile("ip_" + fileBase)
	if err != nil {
		return nil
	}
	return &classifier{
		PN,
		NI,
		IP,
	}
}

func (c *classifier) Save() {
	pn, err := os.OpenFile("pn_"+fileBase, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	ni, err := os.OpenFile("ni_"+fileBase, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	ip, err := os.OpenFile("ip_"+fileBase, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	c.PN.WriteTo(pn)
	c.NI.WriteTo(ni)
	c.IP.WriteTo(ip)
}

func (c *classifier) Learn(document []string, which bayesian.Class) {
	if which == P {
		c.PN.Learn(document, which)
		c.IP.Learn(document, which)
	} else if which == N {
		c.NI.Learn(document, which)
		c.PN.Learn(document, which)
	} else if which == I {
		c.IP.Learn(document, which)
		c.NI.Learn(document, which)
	}
}

func (c *classifier) DeliberationNP(s string) bayesian.Class {
	var doc = []string{}
	var (
		PScore = float64(1.0)
		NScore = float64(1.0)
		IScore = float64(1.0)
	)
	nodes, err := mecab.Parse(s)
	if err != nil {
		return ""
	}

	for _, node := range nodes {
		doc = append(doc, node.Base)
	}

	pnScores, _, pnb := c.PN.LogScores(doc)
	if pnb {
		PScore = PScore * (-1 * pnScores[0])
		NScore = NScore * (-1 * pnScores[1])
	}
	niScores, _, nib := c.NI.LogScores(doc)
	if nib {
		NScore = NScore * (-1 * niScores[0])
		IScore = IScore * (-1 * niScores[1])
	}
	ipScores, _, ipb := c.IP.LogScores(doc)
	if ipb {
		IScore = IScore * (-1 * ipScores[0])
		PScore = PScore * (-1 * ipScores[1])
	}

	if PScore < NScore {
		if PScore < IScore {
			return P
		}
	} else {
		if NScore < IScore {
			return N
		}
	}
	return I
}

func learn(s string, which bayesian.Class, c *classifier) *classifier {
	var doc = []string{}
	nodes, err := mecab.Parse(s)
	if err != nil {
		return c
	}

	for _, node := range nodes {
		doc = append(doc, node.Base)
	}
	c.Learn(doc, which)
	return c
}

func main() {
	var c *classifier
	var success int
	var param string
	c = load()
	if c == nil {
		c = newClassifier()
	}

	result := c.DeliberationNP(*input)
	fmt.Println(result)

	fmt.Println("正解のとき1, 失敗の時0")
	for {
		_, err := fmt.Scan(&success)
		if err != nil {
			os.Exit(1)
		}
		if success == 1 {
			learn(*input, result, c)
			os.Exit(0)
		} else if success == 0 {
			break
		}
	}

	fmt.Println("p, n, iのいずれかを選ぶ")

	for {
		_, err := fmt.Scan(&param)
		if err != nil {
			os.Exit(1)
		}
		if param == "p" || param == "n" || param == "i" {
			break
		}
	}

	if param == "p" {
		fmt.Println(P)
		learn(*input, P, c)
	} else if param == "n" {
		fmt.Println(N)
		learn(*input, N, c)
	} else if param == "i" {
		fmt.Println(I)
		learn(*input, I, c)
	}

	c.Save()
}
