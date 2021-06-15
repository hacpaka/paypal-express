package paypal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	NVP_SANDBOX_URL         = "https://api-3t.sandbox.paypal.com/nvp"
	NVP_PRODUCTION_URL      = "https://api-3t.paypal.com/nvp"
	CHECKOUT_SANDBOX_URL    = "https://www.sandbox.paypal.com/cgi-bin/webscr"
	CHECKOUT_PRODUCTION_URL = "https://www.paypal.com/cgi-bin/webscr"
	NVP_VERSION             = "84"
)

type PayPalClient struct {
	username string
	password string
	signature string
	usesSandbox bool
	client *http.Client
}

type PayPalDigitalGood struct {
	Name string
	Amount float64
	Quantity int16
}

type PayPalResponse struct {
	Ack string
	CorrelationId string
	Timestamp string
	Version string
	Build string
	Values url.Values
	usedSandbox bool
	Invnum string
	TransactionId string
}

type PayPalError struct {
	Ack string
	ErrorCode string
	ShortMessage string
	LongMessage string
	SeverityCode string
}

func (e *PayPalError) Error() string {
	var message string
	if len(e.ErrorCode) != 0 && len(e.ShortMessage) != 0 {
		message = "PayPal Error " + e.ErrorCode + ": " + e.ShortMessage
	} else if len(e.Ack) != 0 {
		message = e.Ack
	} else {
		message = "PayPal is undergoing maintenance.\nPlease try again later."
	}

	return message
}

func (r *PayPalResponse) CheckoutUrl() string {
	query := url.Values{}
	query.Set("cmd", "_express-checkout")
	query.Add("token", r.Values["TOKEN"][0])
	checkoutUrl := CHECKOUT_PRODUCTION_URL
	if r.usedSandbox {
		checkoutUrl = CHECKOUT_SANDBOX_URL
	}
	return fmt.Sprintf("%s?%s", checkoutUrl, query.Encode())
}

func SumPayPalDigitalGoodAmounts(goods *[]PayPalDigitalGood) (sum float64) {
	for _, dg := range *goods {
		sum += dg.Amount * float64(dg.Quantity)
	}
	return
}

func NewDefaultClient(username, password, signature string, usesSandbox bool) *PayPalClient {
	return &PayPalClient{username, password, signature, usesSandbox, new(http.Client)}
}

func NewClient(username, password, signature string, usesSandbox bool, client *http.Client) *PayPalClient {
	return &PayPalClient{username, password, signature, usesSandbox, client}
}

func (pClient *PayPalClient) PerformRequest(values url.Values) (*PayPalResponse, error) {
	values.Add("USER", pClient.username)
	values.Add("PWD", pClient.password)
	values.Add("SIGNATURE", pClient.signature)
	values.Add("VERSION", NVP_VERSION)

	endpoint := NVP_PRODUCTION_URL
	if pClient.usesSandbox {
		endpoint = NVP_SANDBOX_URL
	}

	formResponse, err := pClient.client.PostForm(endpoint, values)
	if err != nil {
		return nil, err
	}
	defer formResponse.Body.Close()

	body, err := ioutil.ReadAll(formResponse.Body)
	if err != nil {
		return nil, err
	}

	responseValues, err := url.ParseQuery(string(body))
	response := &PayPalResponse{usedSandbox: pClient.usesSandbox}
	if err == nil {
		response.Ack = responseValues.Get("ACK")
		response.CorrelationId = responseValues.Get("CORRELATIONID")
		response.Timestamp = responseValues.Get("TIMESTAMP")
		response.Version = responseValues.Get("VERSION")
		response.Build = responseValues.Get("2975009")
		response.Values = responseValues
		response.Invnum = responseValues.Get("PAYMENTREQUEST_0_INVNUM")
		response.TransactionId = responseValues.Get("PAYMENTREQUEST_0_TRANSACTIONID")

		errorCode := responseValues.Get("L_ERRORCODE0")
		if len(errorCode) != 0 || strings.ToLower(response.Ack) == "failure" || strings.ToLower(response.Ack) == "failurewithwarning" {
			pError := new(PayPalError)
			pError.Ack = response.Ack
			pError.ErrorCode = errorCode
			pError.ShortMessage = responseValues.Get("L_SHORTMESSAGE0")
			pError.LongMessage = responseValues.Get("L_LONGMESSAGE0")
			pError.SeverityCode = responseValues.Get("L_SEVERITYCODE0")

			err = pError
		}
	}

	return response, err
}


/*
map[
ACK:[Success]
BUILD:[55698494]
CORRELATIONID:[9b81bd1f79546]
INSURANCEOPTIONSELECTED:[false]
PAYMENTINFO_0_ACK:[Success]
PAYMENTINFO_0_AMT:[1.00]
PAYMENTINFO_0_CURRENCYCODE:[USD]
PAYMENTINFO_0_ERRORCODE:[0]
PAYMENTINFO_0_FEEAMT:[0.33]
PAYMENTINFO_0_ORDERTIME:[2021-06-14T07:34:08Z]
PAYMENTINFO_0_PAYMENTSTATUS:[Completed]
PAYMENTINFO_0_PAYMENTTYPE:[instant]
PAYMENTINFO_0_PENDINGREASON:[None]
PAYMENTINFO_0_PROTECTIONELIGIBILITY:[Eligible]
PAYMENTINFO_0_PROTECTIONELIGIBILITYTYPE:[ItemNotReceivedEligible,UnauthorizedPaymentEligible]
PAYMENTINFO_0_REASONCODE:[None]
PAYMENTINFO_0_SECUREMERCHANTACCOUNTID:[P2DAAH6XJ5996]
PAYMENTINFO_0_SELLERPAYPALACCOUNTID:[sb-uuum436476396@business.example.com]
PAYMENTINFO_0_TAXAMT:[0.00]
PAYMENTINFO_0_TRANSACTIONID:[8WB29409YN938500U]
PAYMENTINFO_0_TRANSACTIONTYPE:[expresscheckout]
SHIPPINGOPTIONISDEFAULT:[false]
SUCCESSPAGEREDIRECTREQUESTED:[false]
TIMESTAMP:[2021-06-14T07:34:08Z]
TOKEN:[EC-3A590955NS779441R]
VERSION:[84]] true}
*/


/*
map[
ACK:[Success]
ADDRESSSTATUS:[Confirmed]
AMT:[1.00]
BILLINGAGREEMENTACCEPTEDSTATUS:[0]
BUILD:[55698494]
CHECKOUTSTATUS:[PaymentActionCompleted]
CORRELATIONID:[f0b2efe209c62]
COUNTRYCODE:[US]
CURRENCYCODE:[USD]
EMAIL:[sb-47awgh6155709@personal.example.com]
FIRSTNAME:[John]
HANDLINGAMT:[0.00]
INSURANCEAMT:[0.00]
INSURANCEOPTIONOFFERED:[false]
INVNUM:[5d617644e6324c29bd3de7718e35232d]
LASTNAME:[Doe]
PAYERID:[H97ABK6WNY7NU]
PAYERSTATUS:[verified]
PAYMENTREQUESTINFO_0_ERRORCODE:[0]
PAYMENTREQUESTINFO_0_TRANSACTIONID:[5X372240SX0980714]
PAYMENTREQUEST_0_AMT:[1.00]
PAYMENTREQUEST_0_CURRENCYCODE:[USD]
PAYMENTREQUEST_0_HANDLINGAMT:[0.00]
PAYMENTREQUEST_0_INSURANCEAMT:[0.00]
PAYMENTREQUEST_0_INSURANCEOPTIONOFFERED:[false]
PAYMENTREQUEST_0_INVNUM:[5d617644e6324c29bd3de7718e35232d]
PAYMENTREQUEST_0_SELLERPAYPALACCOUNTID:[sb-uuum436476396@business.example.com]
PAYMENTREQUEST_0_SHIPDISCAMT:[0.00]
PAYMENTREQUEST_0_SHIPPINGAMT:[0.00]
PAYMENTREQUEST_0_TAXAMT:[0.00]
PAYMENTREQUEST_0_TRANSACTIONID:[5X372240SX0980714]
SHIPDISCAMT:[0.00]
SHIPPINGAMT:[0.00]
TAXAMT:[0.00]
TIMESTAMP:[2021-06-14T07:48:18Z]
TOKEN:[EC-4AJ496724V5128253]
TRANSACTIONID:[5X372240SX0980714]
VERSION:[84]] true}
 */

func (pClient *PayPalClient) SetExpressCheckoutDigitalGoods(paymentAmount float64, currencyCode string, returnURL, cancelURL string, invnum string, goods []PayPalDigitalGood) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "SetExpressCheckout")
	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", paymentAmount))
	values.Add("PAYMENTREQUEST_0_PAYMENTACTION", "Sale")
	values.Add("PAYMENTREQUEST_0_CURRENCYCODE", currencyCode)
	values.Add("PAYMENTREQUEST_0_INVNUM", invnum)
	values.Add("RETURNURL", returnURL)
	values.Add("CANCELURL", cancelURL)
	values.Add("REQCONFIRMSHIPPING", "0")
	values.Add("NOSHIPPING", "1")
	values.Add("SOLUTIONTYPE", "Sole")

	for i := 0; i < len(goods); i++ {
		good := goods[i]

		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_NAME", i), good.Name)
		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_AMT", i), fmt.Sprintf("%.2f", good.Amount))
		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_QTY", i), fmt.Sprintf("%d", good.Quantity))
		values.Add(fmt.Sprintf("%s%d", "L_PAYMENTREQUEST_0_ITEMCATEGORY", i), "Digital")
	}

	return pClient.PerformRequest(values)
}

// Convenience function for Sale (Charge)
func (pClient *PayPalClient) DoExpressCheckoutSale(token, payerId, currencyCode string, finalPaymentAmount float64) (*PayPalResponse, error) {
	return pClient.DoExpressCheckoutPayment(token, payerId, "Sale", currencyCode, finalPaymentAmount)
}

// paymentType can be "Sale" or "Authorization" or "Order" (ship later)
func (pClient *PayPalClient) DoExpressCheckoutPayment(token, payerId, paymentType, currencyCode string, finalPaymentAmount float64) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "DoExpressCheckoutPayment")
	values.Add("TOKEN", token)
	values.Add("PAYERID", payerId)
	values.Add("PAYMENTREQUEST_0_PAYMENTACTION", paymentType)
	values.Add("PAYMENTREQUEST_0_CURRENCYCODE", currencyCode)
	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", finalPaymentAmount))

	return pClient.PerformRequest(values)
}

func (pClient *PayPalClient) GetExpressCheckoutDetails(token string) (*PayPalResponse, error) {
	values := url.Values{}
	values.Add("TOKEN", token)
	values.Set("METHOD", "GetExpressCheckoutDetails")
	return pClient.PerformRequest(values)
}
