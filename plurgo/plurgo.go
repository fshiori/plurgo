package plurgo

import (
    "encoding/json"
    "fmt"
    "github.com/garyburd/go-oauth/oauth"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    //"time"
)

type PlurkCredentials struct {
    ConsumerToken string
    ConsumerSecret string
    AccessToken string
    AccessSecret string
}

var baseURL = "http://www.plurk.com"

var oauthClient = oauth.Client {
    TemporaryCredentialRequestURI: "http://www.plurk.com/OAuth/request_token",
    ResourceOwnerAuthorizationURI: "http://www.plurk.com/OAuth/authorize",
    TokenRequestURI:               "http://www.plurk.com/OAuth/access_token",
}

var plurkOAuth PlurkCredentials
var signinOAuthClient oauth.Client

func ReadCredentials(credPath string) (*PlurkCredentials, error) {
    b, err := ioutil.ReadFile(credPath)
    if err != nil {
	return nil, err
    }
    var cred PlurkCredentials
    err = json.Unmarshal(b, &cred)
    if err != nil {
	return nil, err
    }
    return &cred, nil
}

func doAuth(requestToken *oauth.Credentials) (*oauth.Credentials, error) {
    _url := oauthClient.AuthorizationURL(requestToken, nil)
    fmt.Println("Open the following URL and authorize it:", _url)

    var pinCode string
    fmt.Print("Input the PIN code: ")
    fmt.Scan(&pinCode)
    accessToken, _, err := oauthClient.RequestToken(http.DefaultClient, requestToken, pinCode)
    if err != nil {
	log.Fatal("failed to request token:", err)
    }
    return accessToken, nil
}

func GetAccessToken(cred *PlurkCredentials) (*oauth.Credentials, bool, error) {
    oauthClient.Credentials.Token = cred.ConsumerToken
    oauthClient.Credentials.Secret = cred.ConsumerSecret

    authorized := false
    var token *oauth.Credentials
    if cred.AccessToken != "" && cred.AccessSecret != "" {
	token = &oauth.Credentials{cred.AccessToken, cred.AccessSecret}
    } else {
	requestToken, err := oauthClient.RequestTemporaryCredentials(http.DefaultClient, "", nil)
	if err != nil {
	    log.Printf("failed to request temporary credentials: %v", err)
	    return nil, false, err
	}
	token, err = doAuth(requestToken)
	if err != nil {
	    log.Printf("failed to request temporary credentials: %v", err)
	    return nil, false, err
	}

	cred.AccessToken = token.Token
	cred.AccessSecret = token.Secret
	authorized = true
    }
    return token, authorized, nil
}

func CallAPI(token *oauth.Credentials, _url string, opt map[string]string) ([]byte, error) {
    var apiURL = baseURL + _url
    param := make(url.Values)
    for k, v := range opt {
	param.Set(k, v)
    }
    oauthClient.SignParam(token, "POST", apiURL, param)
    res, err := http.PostForm(apiURL, url.Values(param))
    if err != nil {
	log.Println("failed to call API:", err)
	return nil, err
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
	log.Println("failed to call API:", err)
	return nil, err
    }
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
	log.Println("failed to get response:", err)
	return nil, err
    }
    return body, nil
}
