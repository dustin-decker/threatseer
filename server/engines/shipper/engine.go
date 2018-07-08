package shipper

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/dustin-decker/threatseer/server/config"
	"github.com/dustin-decker/threatseer/server/models"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	_ "github.com/jackc/pgx/stdlib" // postgres driver for the ORM
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

// Shipper makes it compatible flow pipeline
type Shipper struct {
	b      *beat.Beat
	beat   beat.Client
	db     orm.Ormer
	dbChan chan models.Event
	config config.Config
}

// PublishFromPipeline is the entrypoint from the flow pipeline
func (s *Shipper) PublishFromPipeline(in chan models.Event) {

	for e := range in {
		var riskScore int
		for _, indicator := range e.Indicators {
			riskScore = riskScore + indicator.Score
		}
		evnt := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"event":      e.Event,
				"indicators": e.Indicators,
				"risk_score": riskScore,
				"src_ip":     e.ClientAddr,
			},
		}

		if s.config.BeatsOutput {
			// goes to beats output
			s.beat.Publish(evnt)
		}

		if s.config.PostgresOutput {
			// goes to DB output
			s.dbChan <- e
		}
	}
}

func (b *batcher) insertBatches(minSize int) {
	startTime := time.Now()
	var num int64

	if len(b.indicatorBatch) > minSize {
		numEvents, err := b.db.InsertMulti(b.batchSize, b.indicatorBatch)
		if err != nil {
			log.Fatal("got error during bulk insert: ", err)
		}
		num += numEvents
	}
	if len(b.processEventsBatch) > minSize {
		numEvents, err := b.db.InsertMulti(b.batchSize, b.processEventsBatch)
		if err != nil {
			log.Fatal("got error during bulk insert: ", err)
		}
		num += numEvents
	}
	if len(b.kernelEventsBatch) > minSize {
		numEvents, err := b.db.InsertMulti(b.batchSize, b.kernelEventsBatch)
		if err != nil {
			log.Fatal("got error during bulk insert: ", err)
		}
		num += numEvents
	}

	var duration = time.Since(startTime)

	log.WithFields(log.Fields{
		"num_events": num,
		"elasped":    duration.String(),
	}).Info("inserted events into database")
}

type batcher struct {
	db                 orm.Ormer
	batchSize          int
	indicatorBatch     []models.Indicator
	processEventsBatch []models.ProcessEvent
	kernelEventsBatch  []models.KernelEvent
}

func (b *batcher) reset() {
	b.indicatorBatch = []models.Indicator{}
	b.processEventsBatch = []models.ProcessEvent{}
	b.kernelEventsBatch = []models.KernelEvent{}
}

// batchInsertEventsIntoDB batcher events up to max batch size or max duration
// One output for all pipelines.
func (s *Shipper) batchInsertEventsIntoDB(in <-chan models.Event) {
	b := batcher{db: s.db, batchSize: s.config.PostgresBatchSize}

	go func() {
		i := 0
		tick := time.Tick(1 * time.Second)

		for {
			select {
			// process whatever we have seen so far if the batch size isn't filled in 5 secs
			case <-tick:
				if i > 30 {
					b.insertBatches(10)
					b.reset()
					i = 0
				}
			case e, ok := <-in:
				if !ok {
					break
				}

				// indicators
				for index := range e.Indicators {
					e.Indicators[index].ID = xid.New().String()
					e.Indicators[index].ProcessEventID = e.Event.GetId()
					b.indicatorBatch = append(b.indicatorBatch, e.Indicators[index])
					i++
				}

				// process events
				processEvent := models.GetProcessEventContext(e)
				if processEvent != nil {
					processEvent.ID = e.Event.GetId()
					b.processEventsBatch = append(b.processEventsBatch, *processEvent)
					i++
				}

				// kernel call events
				kernelEvent := models.GetKernelEventContext(e)
				if kernelEvent != nil {
					kernelEvent.ID = e.Event.GetId()
					b.kernelEventsBatch = append(b.kernelEventsBatch, *kernelEvent)
					i++
				}

				if i > b.batchSize {
					b.insertBatches(b.batchSize)
					b.reset()
					i = 0
				}

			}
		}
	}()

}

// NewShipperEngine is the entrypoint for the datashipper
func NewShipperEngine(b *beat.Beat, c config.Config) Shipper {

	s := Shipper{config: c}

	if c.BeatsOutput {
		beatClient, err := b.Publisher.Connect()
		if err != nil {
			log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not connect to shipper publisher")
		}
		s.beat = beatClient
	}

	if c.PostgresOutput {
		err := orm.RegisterDriver("pgx", orm.DRPostgres)
		if err != nil {
			log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not register ORM driver")
		}

		maxIdle := 15
		maxConn := 15
		pgURI := fmt.Sprintf("postgres://threatseer@%s:5432/threatseer", c.PostgresHost)
		err = orm.RegisterDataBase("default", "pgx", pgURI, maxIdle, maxConn)
		if err != nil {
			log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not register database")
		}
		// orm.DefaultTimeLoc = time.UTC
		orm.RegisterModel(
			new(models.ProcessEvent),
			new(models.KernelEvent),
			new(models.Indicator),
		)
		err = orm.RunSyncdb("default", true, true)
		if err != nil {
			log.WithFields(log.Fields{"engine": "shipper", "err": err}).Fatal("could not sync database")
		}

		db := orm.NewOrm()
		db.Using("default")

		inDB := make(chan models.Event, 5000)
		s.db = db
		s.dbChan = inDB
		go s.batchInsertEventsIntoDB(inDB)
	}

	return s
}
