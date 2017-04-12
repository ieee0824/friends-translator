package main

//import (
//	"fmt"
//	"io/ioutil"
//
//	"github.com/JesusIslam/tldr"
//)
//
//func main() {
//	intoSentences := 3
//	textB, _ := ioutil.ReadFile("./sample.txt")
//	text := string(textB)
//	bag := tldr.New()
//	result, _ := bag.Summarize(text, intoSentences)
//	fmt.Println(result)
//}

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	cabocha "github.com/ledyba/go-cabocha"
	mecab "github.com/yukihir0/mecab-go"
)

var (
	input = flag.String("i", "", "")
)

type NPElem struct {
	p int
	s string
}

var index = map[string][]NPElem{}

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
	np, _ := CalcNP(*input)

	if np == 0 {
		fmt.Println("わかんないやー")
		return
	} else if 1 <= np && cw != "" {
		fmt.Println(PosiCon(cw))
		return
	} else if 1 <= np {
		fmt.Println("わたしにはよくわからないけどすごいんだよー")
		return
	}
	fmt.Println(NegaCon(cw))
}
