package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jbrukh/bayesian"
	cabocha "github.com/ledyba/go-cabocha"
	mecab "github.com/yukihir0/mecab-go"
)

var (
	input = flag.String("i", "", "input string")
	mode  = flag.Bool("m", false, "if use bayesian: true")
)

type NPElem struct {
	p int
	s string
}

var index = map[string][]NPElem{}

// classの定義
const (
	P bayesian.Class = "Posi"
	N bayesian.Class = "Nega"
	I bayesian.Class = "Illegal"
)

const fileBase = "nb.dat"

type classifier struct {
	PN *bayesian.Classifier
	NI *bayesian.Classifier
	IP *bayesian.Classifier
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
		log.Println(err)
		return nil
	}
	NI, err := bayesian.NewClassifierFromFile("ni_" + fileBase)
	if err != nil {
		log.Println(err)
		return nil
	}
	IP, err := bayesian.NewClassifierFromFile("ip_" + fileBase)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &classifier{
		PN,
		NI,
		IP,
	}
}

func init() {
	flag.Parse()
	readNPIndex("wago.121808.pn")
}

func PosiCon(s string) string {
	return fmt.Sprintf("すごーい。君は%sフレンズなんだね。\n", s)
}

func NegaCon(s string) string {
	return fmt.Sprintf("だいじょーぶ。フレンズには得意不得意があるからー。")
}

func TrimSubject(s string) (string, error) {
	var ret string
	c := cabocha.MakeCabocha()
	sentence, err := c.Parse(s)
	if err != nil {
		return "", err
	}
	for i, chunk := range sentence.Chunks {
		if i == 0 && chunk.Tokens[0].Features[0] == "名詞" {
			continue
		} else if ret == "" && chunk.Tokens[0].Features[0] == "助詞" {
			continue
		}
		for _, token := range chunk.Tokens {
			ret += token.Body
		}
	}
	return ret, nil
}

//第一引数活用させたもの
//第二引数未活用のもの
func ExtractCharacteristicWords(s string) (string, string, error) {
	nodes, err := mecab.Parse(s)
	if err != nil {
		return "", "", err
	}
	var ret = []string{}
	if err != nil {
		return "", "", err
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		if nodes[i].Pos != "名詞" && nodes[i].Pos != "形容詞" {
			nodes = nodes[:i]
		} else {
			for _, node := range nodes {
				ret = append(ret, node.Surface)
			}

			if nodes[i].Pos1 == "形容動詞語幹" {
				ret = append(ret, "な")
			} else if nodes[i].Pos1 == "一般" {
				ret = append(ret, "の")
			} else if nodes[i].Pos == "形容詞" {
				ret[len(ret)-1] = nodes[i].Base
			}
			break
		}
	}

	if len(ret) == 0 {
		return "", "", errors.New("can not convert")
	}

	return strings.Join(ret, ""), strings.Join(ret[:len(ret)-1], ""), nil
}

func parseNP(s string) int {
	if strings.Contains(s, "ネガ") {
		return -1
	} else if strings.Contains(s, "ポジ") {
		return 1
	}
	return 0
}

func readNPIndex(path string) error {
	var err error
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	for line := ""; err == nil; line, err = reader.ReadString('\n') {
		line = strings.Replace(line, "\n", "", -1)
		i := strings.Split(line, "\t")
		if len(i) < 2 {
			continue
		}
		np := parseNP(i[0])
		words := strings.Split(i[1], " ")

		e := NPElem{np, i[1]}
		key := words[0]
		index[key] = append(index[key], e)
	}
	return nil
}

func CalcNP(s string) (int, error) {
	var result int
	nodes, err := mecab.Parse(s)
	if err != nil {
		return 0, err
	}
	for _, node := range nodes {
		elems, ok := index[node.Base]
		if !ok {
			continue
		}

		result += elems[0].p
	}
	return result, nil
}

func main() {
	s, _ := TrimSubject(*input)
	cw, _, _ := ExtractCharacteristicWords(s)

	if !*mode {
		np, _ := CalcNP(*input)

		if 0 <= np && cw != "" {
			fmt.Println(PosiCon(cw))
			return
		} else if 1 <= np {
			fmt.Println("わたしにはよくわからないけどすごいんだよー")
			return
		}
		fmt.Println(NegaCon(cw))
		return
	}
	c := load()
	if c == nil {
		return
	}

	result := c.DeliberationNP(*input)

	if result == "Posi" {
		fmt.Println(PosiCon(cw))
	} else if result == "Nega" {
		fmt.Println(NegaCon(cw))
	} else {
		fmt.Println("わたしにはよくわからないけどすごいんだよー")
	}
}
