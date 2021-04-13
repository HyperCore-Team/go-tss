package monitor

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Metric struct {
	keygenCounter     *prometheus.CounterVec
	keysignCounter    *prometheus.CounterVec
	keyRegroupCounter *prometheus.CounterVec
	joinPartyCounter  *prometheus.CounterVec
	keySignTime       prometheus.Gauge
	keyGenTime        prometheus.Gauge
	keyReGroupTime    prometheus.Gauge
	joinPartyTime     *prometheus.GaugeVec
	logger            zerolog.Logger
}

func (m *Metric) UpdateKeyGen(keygenTime time.Duration, success bool) {
	if success {
		m.keyGenTime.Set(float64(keygenTime))
		m.keygenCounter.WithLabelValues("success").Inc()
	} else {
		m.keygenCounter.WithLabelValues("failure").Inc()
	}
}

func (m *Metric) UpdateKeySign(keysignTime time.Duration, success bool) {
	if success {
		m.keySignTime.Set(float64(keysignTime))
		m.keysignCounter.WithLabelValues("success").Inc()
	} else {
		m.keysignCounter.WithLabelValues("failure").Inc()
	}
}

func (m *Metric) UpdateKeyRegroup(keygenTime time.Duration, success bool) {
	if success {
		m.keyReGroupTime.Set(float64(keygenTime))
		m.keyRegroupCounter.WithLabelValues("success").Inc()
	} else {
		m.keyRegroupCounter.WithLabelValues("failure").Inc()
	}
}

func (m Metric) KeygenJoinParty(joinpartyTime time.Duration, success bool) {
	if success {
		m.joinPartyTime.WithLabelValues("keygen").Set(float64(joinpartyTime))
		m.joinPartyCounter.WithLabelValues("keygen", "success").Inc()
	} else {
		m.joinPartyCounter.WithLabelValues("keygen", "failure").Inc()
	}
}

func (m *Metric) KeysignJoinParty(joinpartyTime time.Duration, success bool) {
	if success {
		m.joinPartyTime.WithLabelValues("keysign").Set(float64(joinpartyTime))
		m.joinPartyCounter.WithLabelValues("keysign", "success").Inc()
	} else {
		m.joinPartyCounter.WithLabelValues("keysign", "failure").Inc()
	}
}

func (m *Metric) KeyRegroupJoinParty(joinpartyTime time.Duration, success bool) {
	if success {
		m.joinPartyTime.WithLabelValues("keyregroup").Set(float64(joinpartyTime))
		m.joinPartyCounter.WithLabelValues("keyregroup", "success").Inc()
	} else {
		m.joinPartyCounter.WithLabelValues("keyregroup", "failure").Inc()
	}
}

func (m *Metric) Enable() {
	prometheus.MustRegister(m.keygenCounter)
	prometheus.MustRegister(m.keysignCounter)
	prometheus.MustRegister(m.joinPartyCounter)
	prometheus.MustRegister(m.keyGenTime)
	prometheus.MustRegister(m.keySignTime)
	prometheus.MustRegister(m.joinPartyTime)
}

func NewMetric() *Metric {
	metrics := Metric{

		keygenCounter: prometheus.NewCounterVec(

			prometheus.CounterOpts{
				Namespace: "Tss",
				Subsystem: "Tss",
				Name:      "keygen",
				Help:      "Tss keygen success and failure counter",
			},
			[]string{"status"},
		),

		keysignCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "Tss",
				Subsystem: "Tss",
				Name:      "keysign",
				Help:      "Tss keysign success and failure counter",
			},
			[]string{"status"},
		),

		keyRegroupCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "Tss",
				Subsystem: "Tss",
				Name:      "keyregroup",
				Help:      "Tss keyregroup success and failure counter",
			},
			[]string{"status"},
		),

		joinPartyCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "Tss",
			Subsystem: "Tss",
			Name:      "join_party",
			Help:      "Tss keygen join party success and failure counter",
		}, []string{
			"type", "result",
		}),

		keyGenTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "Tss",
				Subsystem: "Tss",
				Name:      "keygen_time",
				Help:      "the time spend for the latest keygen",
			},
		),

		keySignTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "Tss",
				Subsystem: "Tss",
				Name:      "keysign_time",
				Help:      "the time spend for the latest keysign",
			},
		),

		keyReGroupTime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "Tss",
				Subsystem: "Tss",
				Name:      "keyregroup_time",
				Help:      "the time spend for the latest keysign/keygen join party",
			},
		),

		joinPartyTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "Tss",
				Subsystem: "Tss",
				Name:      "joinparty_time",
				Help:      "the time spend for the latest keysign/keygen join party",
			}, []string{"type"}),

		logger: log.With().Str("module", "tssMonitor").Logger(),
	}
	return &metrics
}
