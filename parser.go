package toloka

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

const (
	hUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:104.0) Gecko/20100101 Firefox/104.0"
	host       = "https://toloka.to/"
)

var (
	ErrPageParse = errors.New("page parsing error")
)

type TorrentFile struct {
	Name     string
	FileURL  string
	FileName string
	Magnet   string
	Id       int
}

type Parser struct {
	client *http.Client
}

type LoginData struct {
	Username string
	Password string
}

func (p Parser) do(url string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Referer", url)
	req.Header.Set("User-Agent", hUserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, err
}

func New() Parser {
	client := &http.Client{
		Jar: func() *cookiejar.Jar {
			jar, _ := cookiejar.New(nil)
			return jar
		}(),
	}

	return Parser{
		client: client,
	}
}

func (p *Parser) Login(l LoginData) error {
	loginURL := host + "login.php"

	data := url.Values{
		"username":  {l.Username},
		"password":  {l.Password},
		"autologin": {"on"},
		"ssl":       {"on"},
		"redirect":  {"/"},
		"login":     {"Вхід"},
	}

	resp, err := p.client.PostForm(loginURL, data)
	if err != nil {
		return fmt.Errorf("login error: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (p Parser) Parse(url string) (TorrentFile, error) {
	var torrent TorrentFile
	doc, err := p.do(url)
	if err != nil {
		return TorrentFile{}, err
	}

	torrentTable := doc.Find(".btTbl")
	if torrentTable.Length() == 0 {
		return TorrentFile{}, ErrPageParse
	}

	torrent.FileURL, _ = doc.Find(".piwik_download").Attr("href")
	torrent.Id, _ = strconv.Atoi(torrent.FileURL[1:])
	torrent.FileURL = host + torrent.FileURL

	torrent.Magnet, _ = torrentTable.Find(".gensmall > a").Attr("href")
	torrent.FileName = torrentTable.Find(".row6_to b").Text()
	torrent.Name = doc.Find(".maintitle").Text()

	return torrent, nil
}

func (p Parser) Search(keywords string) ([]string, error) {
	const defaultParams = "?prev_sd=0&prev_a=0&prev_my=0&prev_n=0&prev_shc=0&prev_shf=1&prev_sha=1&prev_cg=1&prev_ct=1&prev_at=1&prev_nt=1&prev_de=1&prev_nd=1&prev_tcs=1&prev_shs=0&f[]=-1&o=1&s=2&tm=-1&cg=1&ct=1&at=1&nt=1&de=1&nd=1&shf=1&sha=1&tcs=1&sns=-1&sds=-1&pn=&send=Пошук"
	const searchFile = "tracker.php"
	params := defaultParams + "&nm=" + url.QueryEscape(keywords)

	searchURL := host + searchFile + params

	doc, err := p.do(searchURL)
	if err != nil {
		return nil, err
	}

	var links []string
	doc.Find(".prow1, .prow2").Find(".topictitle > a").Each(func(i int, s *goquery.Selection) {
		links = append(links, host+s.AttrOr("href", ""))
	})

	return links, nil
}

func (p Parser) TorrentReader(t TorrentFile) (io.Reader, error) {
	resp, err := p.client.Get(t.FileURL)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
