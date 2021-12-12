package usecase

import (
	"database/sql"
	"github.com/bxcodec/faker/v3"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/widmogrod/software-architecture-playground/datasim/essence/algebra"
	"github.com/widmogrod/software-architecture-playground/datasim/essence/algebra/store"
	"github.com/widmogrod/software-architecture-playground/datasim/essence/algebra/store/postgresql"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type GenerateData struct {
	ID string `faker:"uuid_digit"`
	//Latitude           float32 `faker:"lat"`
	//Longitude          float32 `faker:"long"`
	CreditCardNumber string `faker:"cc_number"`
	CreditCardType   string `faker:"cc_type"`
	Email            string `faker:"email"`
	//DomainName         string  `faker:"domain_name"`
	//IPV4               string  `faker:"ipv4"`
	//IPV6               string  `faker:"ipv6"`
	//Password           string  `faker:"password"`
	//Jwt                string  `faker:"jwt"`
	//PhoneNumber        string  `faker:"phone_number"`
	//MacAddress         string  `faker:"mac_address"`
	//URL                string  `faker:"url"`
	//UserName           string  `faker:"username"`
	//TollFreeNumber     string  `faker:"toll_free_number"`
	//E164PhoneNumber    string  `faker:"e_164_phone_number"`
	//TitleMale          string  `faker:"title_male"`
	//TitleFemale        string  `faker:"title_female"`
	//FirstName          string  `faker:"first_name"`
	//FirstNameMale      string  `faker:"first_name_male"`
	//FirstNameFemale    string  `faker:"first_name_female"`
	//LastName           string  `faker:"last_name"`
	//Name               string  `faker:"name"`
	//UnixTime           int64   `faker:"unix_time"`
	//Date               string  `faker:"date"`
	//Time               string  `faker:"time"`
	//MonthName          string  `faker:"month_name"`
	//Year               string  `faker:"year"`
	//DayOfWeek          string  `faker:"day_of_week"`
	//DayOfMonth         string  `faker:"day_of_month"`
	//Timestamp          string  `faker:"timestamp"`
	//Century            string  `faker:"century"`
	//TimeZone           string  `faker:"timezone"`
	//TimePeriod         string  `faker:"time_period"`
	//Word               string  `faker:"word"`
	//Sentence           string  `faker:"sentence"`
	//Paragraph          string  `faker:"paragraph"`
	//Currency           string  `faker:"currency"`
	//Amount             float64 `faker:"amount"`
	//AmountWithCurrency string  `faker:"amount_with_currency"`
	//UUIDHypenated      string  `faker:"uuid_hyphenated"`
	//UUID               string  `faker:"uuid_digit"`
	//Skip               string  `faker:"-"`
	//PaymentMethod      string  `faker:"oneof: cc, paypal, check, money order"` // oneof will randomly pick one of the comma-separated values supplied in the tag
	//AccountID          int     `faker:"oneof: 15, 27, 61"`                     // use commas to separate the values for now. Future support for other separator characters may be added
	//Price32            float32 `faker:"oneof: 4.95, 9.99, 31997.97"`
	//Price64            float64 `faker:"oneof: 47463.9463525, 993747.95662529, 11131997.978767990"`
	//NumS64             int64   `faker:"oneof: 1, 2"`
	//NumS32             int32   `faker:"oneof: -3, 4"`
	//NumS16             int16   `faker:"oneof: -5, 6"`
	//NumS8              int8    `faker:"oneof: 7, -8"`
	//NumU64             uint64  `faker:"oneof: 9, 10"`
	//NumU32             uint32  `faker:"oneof: 11, 12"`
	//NumU16             uint16  `faker:"oneof: 13, 14"`
	//NumU8              uint8   `faker:"oneof: 15, 16"`
	//NumU               uint    `faker:"oneof: 17, 18"`
	Typ string `faker:"oneof: customer, customer"`
}

var (
	rows int = 10
)

func TestCreateData(t *testing.T) {
	data := GenerateData{
		// To have predictable range of keys to read
		ID: strconv.Itoa(23123),
	}

	relation := CreateShape(data)
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	assert.NoError(t, err)
	s := postgresql.NewStore(db, relation)
	err = s.InitiateShape()
	assert.NoError(t, err)

	//s := store.NewStore()
	metrics := algebra.NewDataCollector()

	metrics.Push("experiment.name", "TestCreateData", algebra.Metadata{
		"runtime.runner": "go-test",
		"runtime.gc_on":  "yes",
	})

	// TODO
	// Test for concurrent writes

	for i := 0; i < rows; i++ {
		data := GenerateData{
			// To have predictable range of keys to read
			ID: strconv.Itoa(i),
		}

		err := faker.FakeData(&data)
		assert.NoError(t, err)

		forEachField(data, func(key string, value interface{}) {
			startT := time.Now()
			err = s.Set(data.ID, data.Typ, key, value)
			assert.NoError(t, err)
			endT := time.Now().Sub(startT)

			status := "ok"
			if err != nil {
				status = "err"
			}

			metrics.Push("metric.execution_time_milliseconds", strconv.FormatInt(endT.Milliseconds(), 10), algebra.Metadata{
				"store.operation_name":   "store.Set",
				"store.operation_result": status,
				"store.value_has_type":   reflect.ValueOf(value).Kind().String(),
				"store.key_exact_name":   key,
			})
		})
	}

	err = metrics.FlushCSV("out.csv")
	assert.NoError(t, err)
}

func CreateShape(data interface{}) store.PrimaryWithMany {
	relation := store.PrimaryWithMany{
		Primary: store.Entity{
			Name:       "customer",
			Attributes: MapToAttributes(data),
		},
		Secondaries: []store.Shape{
			store.Entity{
				Name:       "customer_duplicate",
				Attributes: MapToAttributes(data),
			},
		},
	}

	forEachField(data, func(key string, value interface{}) {

	})

	return relation
}

func MapToAttributes(o interface{}) []store.Attr {
	result := []store.Attr{}
	forEachFieldName(o, func(key string) {
		result = append(result, store.Attribute{
			Name: key,
			// TODO more types, currently corced as string
			Type: store.String{},
		})
	})
	return result
}

func TestRandomRead(t *testing.T) {
	data := GenerateData{
		// To have predictable range of keys to read
		ID: strconv.Itoa(23123),
	}

	relation := CreateShape(data)
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	assert.NoError(t, err)
	s := postgresql.NewStore(db, relation)
	err = s.InitiateShape()
	assert.NoError(t, err)

	//s := store.NewStore()
	metrics := algebra.NewDataCollector()

	metrics.Push("experiment.name", "TestCreateData", algebra.Metadata{
		"runtime.runner": "go-test",
		"runtime.gc_on":  "yes",
	})

	attributes := s.GetAttributes("customer")
	if len(attributes) == 0 {
		attributes = store.AttrList{"name", "no_orders", "last_login_ip"}
	}

	// TODO measure how time changes
	// for different proportions of reads for keys that don't exits
	// for repeating reads
	// for bytes transferred
	// for concurrent reads

	for i := 0; i < rows; i++ {
		startT := time.Now()
		_, err := s.Get(strconv.Itoa(i), "customer", attributes)
		endT := time.Now().Sub(startT)

		status := "ok"
		if err != nil {
			status = "err"
		}

		metrics.Push("metric.execution_time_milliseconds", strconv.FormatInt(endT.Milliseconds(), 10), algebra.Metadata{
			"store.operation_name":   "store.Get",
			"store.operation_result": status,
		})
	}

	err = metrics.FlushCSV("out2.csv")
	assert.NoError(t, err)
}

func forEachField(obj interface{}, handle func(key string, value interface{})) {
	v := reflect.ValueOf(obj)

	for i := 0; i < v.NumField(); i++ {
		handle(v.Type().Field(i).Name, v.Field(i).Interface())
	}
}

func forEachFieldName(obj interface{}, handle func(key string)) {
	v := reflect.ValueOf(obj)

	for i := 0; i < v.NumField(); i++ {
		handle(v.Type().Field(i).Name)
	}
}
