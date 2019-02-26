package client

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type Team struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Log struct {
	Owner      string `json:"owner"`
	CreateDate string `json:"createdDate"`
	Log        string `json:"log"`
}

type ResultWithoutDataField struct {
	ResultMetadata
	Result string `json:"result"`
}

type aResultDoesNotWantDataFieldsToBeParsed struct {
	ResultMetadata
	Logs   []Log  `json:"logs"`
	Offset string `json:"offset"`
}

type aResultWantsDataFieldsToBeParsed struct {
	ResultMetadata
	Teams []Team `json:"data"`
}

func TestParsingWithDataField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `
			{
    "data": [
        {
            "id": "1",
            "name": "n1",
            "description": "d1"
        },
        {
            "id": "2",
            "name": "n2",
            "description": "d2"
        },
        {
            "id": "3",
            "name": "n3",
            "description": "d3"
        }
    ],
    "took": 1.08,
    "requestId": "123"
}
		`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{})
	assert.Equal(t, err.Error(), errors.New("API key cannot be blank").Error())

	ogClient, err = NewOpsGenieClient(&Config{
		ApiKey: "apiKey",
	})
	assert.Nil(t, err)

	request := &testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &aResultWantsDataFieldsToBeParsed{}
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	err = ogClient.Exec(nil, request, result)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, result.Teams[0], Team{Id: "1", Name: "n1", Description: "d1"})
	assert.Equal(t, result.Teams[1], Team{Id: "2", Name: "n2", Description: "d2"})
	assert.Equal(t, result.Teams[2], Team{Id: "3", Name: "n3", Description: "d3"})
}

func TestParsingWithoutDataField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `
			{
    "data": {
        "offset": "123",
        "logs": [
            {
                "owner": "o1",
                "createdDate": "c1",
                "log": "l1"
            },
            {
                "owner": "o2",
                "createdDate": "c2",
                "log": "l2"
            }
        ]
    },
    "took": 0.041,
    "requestId": "123"
}
		`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey: "apiKey",
	})

	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &aResultDoesNotWantDataFieldsToBeParsed{}
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	err = ogClient.Exec(nil, &request, result)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, result.Logs[0], Log{Owner: "o1", CreateDate: "c1", Log: "l1"})
	assert.Equal(t, result.Logs[1], Log{Owner: "o2", CreateDate: "c2", Log: "l2"})
	assert.Equal(t, result.Offset, "123")
}

func TestParsingWhenApiDoesNotReturnDataField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `
			{
				"result": "processed",
				"requestId": "123",
				"took": 0.1
			}
		`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey: "apiKey",
	})

	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &ResultWithoutDataField{}
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	err = ogClient.Exec(nil, &request, result)
	if err != nil {
		t.Fail()
	}
	assert.Equal(t, "processed", result.Result)
}

var (
	BaseURL     = "https://api.opsgenie.com"
	Endpoint    = "v2/alerts"
	EndpointURL = BaseURL + "/" + Endpoint
	BadEndpoint = ":"
)

type testRequest struct {
	BaseRequest
	MandatoryField string
	ExtraField     string
}

func (tr testRequest) Validate() error {
	if tr.MandatoryField == "" {
		return errors.New("mandatory field cannot be empty")
	}

	return nil
}

func (tr testRequest) ResourcePath() string {
	return "/an-enpoint"
}

func (tr testRequest) Method() string {
	return "POST"
}

type testResult struct {
	ResultMetadata
	Data string
}

func TestExec(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
    		"Data": "processed"}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey: "apiKey",
	})

	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	err = ogClient.Exec(nil, &request, result)
	assert.Equal(t, result.Data, "processed")
	if err != nil {
		t.Fail()
	}
}

func TestParsingErrorExec(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey: "apiKey",
	})

	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	err = ogClient.Exec(nil, &request, result)
	assert.Contains(t, err.Error(), "Response could not be parsed, unexpected end of JSON input")
}

func TestExecWhenRequestIsNotValid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
    		"Data": "processed"}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey: "apiKey",
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl

	request := testRequest{ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, &request, result)
	assert.Equal(t, err.Error(), "mandatory field cannot be empty")
}

func TestExecWhenApiReturns422(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintln(w, `{
    "message": "Request body is not processable. Please check the errors.",
    "errors": {
        "recipients#type": "Invalid recipient type 'bb'"
    },
    "took": 0.083,
    "requestId": "Id"
}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey: "apiKey",
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, &request, result)
	fmt.Println(err.Error())
	assert.Contains(t, err.Error(), "422")
	assert.Contains(t, err.Error(), "Invalid recipient")

}

func TestExecWhenApiReturns5XX(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{
    "message": "Internal Server Error",
    "took": 0.083,
    "requestId": "6c20ec4e-076a-4422-8d65-7b8ca92067ab"
}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, &request, result)
	fmt.Println(err.Error())
	assert.Contains(t, err.Error(), "Internal Server Error")
	assert.Contains(t, err.Error(), "500")
}

func TestExecWhenApiReturnsRateLimitingDetails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-RateLimit-State", "THROTTLED")
		w.Header().Add("X-RateLimit-Reason", "ACCOUNT")
		w.Header().Add("X-RateLimit-Period-In-Sec", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintln(w, `{
    "message": "TooManyRequests",
    "took": 1,
    "requestId": "rId"
}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := &testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, request, result)
	assert.Contains(t, err.Error(), "rId")
	assert.Equal(t, "THROTTLED", result.ResultMetadata.RateLimitState)
	assert.Equal(t, "ACCOUNT", result.ResultMetadata.RateLimitReason)
	assert.Equal(t, "60", result.ResultMetadata.RateLimitPeriod)
}

func TestSubscription(t *testing.T) {
	subscriber := MetricSubscriber{
		Process: subscriberProcessImpl,
	}
	subscriber.Register(HTTP)
	subscriber.Register(SDK)
	subscriber.Register(API)

	subscriber2 := MetricSubscriber{}
	subscriber2.Register(HTTP)

	expectedSubsMap := map[string][]MetricSubscriber{
		string(HTTP): {subscriber, subscriber2},
		string(SDK):  {subscriber},
		string(API):  {subscriber},
	}

	assert.Equal(t, len(expectedSubsMap["http"]), len(metricPublisher.SubscriberMap["http"]))
	assert.Equal(t, len(expectedSubsMap["sdk"]), len(metricPublisher.SubscriberMap["sdk"]))
	assert.Equal(t, len(expectedSubsMap["api"]), len(metricPublisher.SubscriberMap["api"]))
}

func subscriberProcessImpl(metric Metric) interface{} {
	return metric
}

func TestHttpMetric(t *testing.T) {
	var httpMetric *HttpMetric
	subscriber := MetricSubscriber{
		Process: func(metric Metric) interface{} {
			httpMetric, _ = metric.(*HttpMetric)
			return httpMetric
		},
	}
	subscriber.Register(HTTP)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{
    "message": "success",
    "took": 1,
    "requestId": "rId"
}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := &testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, request, result)
	assert.Nil(t, err)

	expectedMetric := &HttpMetric{
		RetryCount:   0,
		Error:        nil,
		ResourcePath: "/an-enpoint",
		Status:       "200 OK",
		StatusCode:   200,
	}

	assert.Equal(t, expectedMetric.StatusCode, httpMetric.StatusCode)
	assert.Equal(t, expectedMetric.Status, httpMetric.Status)
	assert.Equal(t, expectedMetric.RetryCount, httpMetric.RetryCount)
	assert.Equal(t, expectedMetric.ResourcePath, httpMetric.ResourcePath)
	assert.Nil(t, httpMetric.Error)
}

func TestHttpMetricWhenRequestRetried(t *testing.T) {
	var httpMetric *HttpMetric
	subscriber := MetricSubscriber{
		Process: func(metric Metric) interface{} {
			httpMetric, _ = metric.(*HttpMetric)
			return httpMetric
		},
	}
	subscriber.Register(HTTP)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintln(w, `{
    "message": "something went wrong",
}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(&Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := &testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, request, result)

	expectedMetric := &HttpMetric{
		RetryCount:   2,
		Error:        err,
		ResourcePath: "/an-enpoint",
		Status:       "504 Gateway Timeout",
		StatusCode:   504,
	}

	assert.Equal(t, expectedMetric.StatusCode, httpMetric.StatusCode)
	assert.Equal(t, expectedMetric.Status, httpMetric.Status)
	assert.Equal(t, expectedMetric.RetryCount, httpMetric.RetryCount)
	assert.Equal(t, expectedMetric.ResourcePath, httpMetric.ResourcePath)
	assert.Nil(t, httpMetric.Error)
}

func TestApiMetric(t *testing.T) {
	var apiMetric *ApiMetric
	subscriber := MetricSubscriber{
		Process: func(metric Metric) interface{} {
			apiMetric, _ = metric.(*ApiMetric)
			return apiMetric
		},
	}
	subscriber.Register(API)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-RateLimit-State", "THROTTLED")
		w.Header().Add("X-RateLimit-Reason", "ACCOUNT")
		w.Header().Add("X-RateLimit-Period-In-Sec", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintln(w, `{
    "message": "TooManyRequests",
    "took": 1,
    "requestId": "rId"
}`)
	}))
	defer ts.Close()

	ogClient, _ := NewOpsGenieClient(&Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := &testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	ogClient.Exec(nil, request, result)
	expectedMetric := ApiMetric{
		ResourcePath: "/an-enpoint",
		ResultMetadata: ResultMetadata{
			RequestId:       "rId",
			ResponseTime:    1,
			RateLimitState:  "THROTTLED",
			RateLimitReason: "ACCOUNT",
			RateLimitPeriod: "60",
			RetryCount:      2,
		},
	}

	assert.Equal(t, expectedMetric.ResourcePath, apiMetric.ResourcePath)
	assert.Equal(t, expectedMetric.ResultMetadata, apiMetric.ResultMetadata)
}

func TestSdkMetricWhenRequestIsNotValid(t *testing.T) {
	var sdkMetric *SdkMetric
	subscriber := MetricSubscriber{
		Process: func(metric Metric) interface{} {
			sdkMetric, _ = metric.(*SdkMetric)
			return sdkMetric
		},
	}
	subscriber.Register(SDK)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintln(w, `{
    "message": "invalid request",
    "took": 1,
    "requestId": "rId"
}`)
	}))
	defer ts.Close()

	ogClient, _ := NewOpsGenieClient(&Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := &testRequest{ExtraField: "extra"}
	result := &testResult{}

	ogClient.Exec(nil, request, result)
	expectedMetric := &SdkMetric{
		ErrorType:         "request-validation-error",
		ErrorMessage:      "mandatory field cannot be empty",
		ResourcePath:      "/an-enpoint",
		SdkRequestDetails: request,
		SdkResultDetails:  result,
	}

	assert.Equal(t, expectedMetric.ResourcePath, sdkMetric.ResourcePath)
	assert.Equal(t, expectedMetric.ErrorType, sdkMetric.ErrorType)
	assert.Equal(t, expectedMetric.ErrorMessage, sdkMetric.ErrorMessage)
	assert.Equal(t, expectedMetric.SdkRequestDetails, sdkMetric.SdkRequestDetails)
	assert.Equal(t, expectedMetric.SdkResultDetails, sdkMetric.SdkResultDetails)
}

func TestSdkMetricWhenExecSuccessful(t *testing.T) {
	var sdkMetric *SdkMetric
	subscriber := MetricSubscriber{
		Process: func(metric Metric) interface{} {
			sdkMetric, _ = metric.(*SdkMetric)
			return sdkMetric
		},
	}
	subscriber.Register(SDK)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{
    "message": "invalid request",
    "took": 1,
    "requestId": "rId"
}`)
	}))
	defer ts.Close()

	ogClient, _ := NewOpsGenieClient(&Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	localUrl := strings.Replace(ts.URL, "http://", "", len(ts.URL)-1)
	ogClient.Config.apiUrl = localUrl
	request := &testRequest{MandatoryField: "f1", ExtraField: "extra"}
	result := &testResult{}

	ogClient.Exec(nil, request, result)
	expectedMetric := &SdkMetric{
		ErrorType:         "",
		ErrorMessage:      "",
		ResourcePath:      "/an-enpoint",
		SdkRequestDetails: request,
		SdkResultDetails:  result,
	}

	assert.Equal(t, expectedMetric.ResourcePath, sdkMetric.ResourcePath)
	assert.Equal(t, expectedMetric.ErrorType, sdkMetric.ErrorType)
	assert.Equal(t, expectedMetric.ErrorMessage, sdkMetric.ErrorMessage)
	assert.Equal(t, expectedMetric.SdkRequestDetails, sdkMetric.SdkRequestDetails)
	assert.Equal(t, expectedMetric.SdkResultDetails, sdkMetric.SdkResultDetails)
}
