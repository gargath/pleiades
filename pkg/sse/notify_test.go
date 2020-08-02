package sse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var responseLines = []string{
	`:ok`,
	`event: message`,
	`id: [{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
	`data: {"$schema":"/mediawiki/recentchange/1.0.0","meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`,
	``,
	`event: message`,
	`id: [{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`,
	`data: {"$schema":"/mediawiki/recentchange/1.0.0",`,
	`data: "meta":{"uri":"https://he.wikipedia.org/wiki/%D7%AA%D7%91%D7%A0%D7%99%D7%AA:%D7%A0%D7%AA%D7%95%D7%A0%D7%99_%D7%9E%D7%93%D7%99%D7%A0%D7%95%D7%AA/%D7%A1%D7%9C%D7%95%D7%91%D7%A7%D7%99%D7%94","request_id":"e386ef4b-75f4-46e8-be93-8f3683d30049","id":"9bea80f8-f99c-4b56-93c4-0eb4272bbcb9","dt":"2020-07-31T14:58:47Z","domain":"he.wikipedia.org","stream":"mediawiki.recentchange","topic":"eqiad.mediawiki.recentchange","partition":0,"offset":2603659077},"id":53404707,"type":"edit","namespace":10,"title":"תבנית:נתוני מדינות/סלובקיה","comment":"bot","timestamp":1596207527,"user":"DMbotY","bot":true,"minor":true,"patrolled":true,"length":{"old":4905,"new":4905},"revision":{"old":28682248,"new":28826355},"server_url":"https://he.wikipedia.org","server_name":"he.wikipedia.org","server_script_path":"/w","wiki":"hewiki","parsedcomment":"bot"}`,
}

var _ = Describe("SSE Consumer", func() {

	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			for _, l := range responseLines {
				fmt.Fprintf(w, l)
				fmt.Fprintf(w, "\n")
			}
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "\n")
		}))
	})

	AfterEach(func() {
		server.Close()
	})

	It("reads and processes events", func() {
		evChan := make(chan *Event)
		clChan := make(chan bool)
		var wg sync.WaitGroup
		events := []Event{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			for e := range evChan {
				events = append(events, *e)
			}
		}()
		go func() {
			defer wg.Done()
			time.Sleep(2 * time.Second)
			close(clChan)
		}()
		err := Notify(server.URL, evChan, clChan)
		close(evChan)
		wg.Wait()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(events)).Should(Equal(2))
		Expect(events[0].Type).Should(Equal("message"))
		Expect(events[1].Type).Should(Equal("message"))
		Expect(events[0].ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
		Expect(events[1].ID).Should(Equal(`[{"topic":"eqiad.mediawiki.recentchange","partition":0,"timestamp":1596207527001},{"topic":"codfw.mediawiki.recentchange","partition":0,"offset":-1}]`))
	})
})