package server_test

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	rs "github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
	"url-shortener/internal/cache"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/route/user/shortener"
	"url-shortener/internal/route/user/sign"
	"url-shortener/internal/server"
)

var _ = Describe("Server APIs", func() {
	var (
		db                     database.MySQLService
		router                 *gin.Engine
		user1                  database.User
		user2                  database.User
		user1AccessTokenHeader string
		user1Url               string
		user1ShortenUrl        string
		user1InvalidUrl        string
	)

	BeforeEach(func() {
		// TODO:
		user1 = database.User{
			Email:    "test5@test5.com",
			Password: "123456",
		}

		user2 = database.User{
			Email:    "test.com",
			Password: "123456",
		}

		user1Url = "https://www.google.com"
		user1InvalidUrl = "hxx://xxx...com"

		env := config.ReadEnv()

		/**
		Database configuration
		*/
		dbConfig := database.Config{
			Username: env.DBUser,
			Password: env.DBPass,
			Host:     env.DBHost,
			Port:     env.DBPort,
		}
		_db, err := database.NewMySQLDatabase(dbConfig)
		if err != nil {
			log.Fatalf("Unable to set up database | Reason: %v\n", err)
		}
		db = _db

		/**
		Caching configuration
		*/
		cache := cache.New(&rs.Options{
			Addr:         fmt.Sprintf("%v:%v", env.RedisHost, env.RedisPort),
			Password:     env.RedisPassword,
			DB:           0,
			ReadTimeout:  time.Minute,
			WriteTimeout: time.Minute,
		})

		jwtKey := []byte(env.JwtKey)
		gConf := sign.GoogleOauthConfig{
			ClientId:     env.GoogleOauthClientId,
			ClientSecret: env.GoogleOauthClientSecret,
		}

		serverOptions := server.ServerOptions{
			Database:                 db,
			Cache:                    cache,
			JwtKey:                   jwtKey,
			UseHttps:                 env.UseHttps,
			BaseUrl:                  env.BaseUrl.String(),
			Domain:                   strings.Split(env.BaseUrl.Host, ":")[0],
			HtmlTemplate:             "../template",
			GoogleOauthConf:          gConf,
			EmailVerificationIgnored: true,
			EmailRequest:             nil,
		}
		router = server.SetupServer(serverOptions)
	})

	Context("Sign up with local account", func() {
		It("should perform successfully", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": "%v"
			}
			`, user1.Email, user1.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusBadRequest)))

			payload = fmt.Sprintf(`
			{
				"email": "%v",
				"code": "123456"
			}
			`, user1.Email)
			recorder = httptest.NewRecorder()
			req = httptest.NewRequest("POST", "/api/user/signup/complete", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusBadRequest)))
		})

		It("should reject the request due to email format problem", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": "%v"
			}
			`, user2.Email, user2.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject the request as email field is empty", func() {
			payload := fmt.Sprintf(`
			{
				"email": "",
				"password": "%v"
			}
			`, user2.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject the request as password field is empty", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": ""
			}
			`, user2.Email)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Sign in with local account", func() {
		It("should perform successfully", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": "%v"
			}
			`, user1.Email, user1.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/sign/", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))

			bodyReader := recorder.Body
			body, err := ioutil.ReadAll(bodyReader)
			Expect(err).ShouldNot(HaveOccurred())
			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			Expect(err).ShouldNot(HaveOccurred())
			accessToken, ok := response["issueToken"]
			Expect(ok).To(Equal(true))
			accessTokenStr, ok := accessToken.(string)
			Expect(ok).To(Equal(true))
			user1AccessTokenHeader = fmt.Sprintf("accessToken=%v", accessTokenStr)
		})

		It("should reject due to field problem", func() {
			payload := fmt.Sprintf(`
			{
				
			}
			`)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/sign/", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Generate a shorten url", func() {
		It("should perform successfully", func() {
			payload := fmt.Sprintf(`
			{
				"url": "%v"
			}
			`, user1Url)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
			bodyReader := recorder.Body
			body, err := ioutil.ReadAll(bodyReader)
			Expect(err).ShouldNot(HaveOccurred())

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			Expect(err).ShouldNot(HaveOccurred())
			url, ok := response["url"]
			Expect(ok).To(Equal(true))
			urlStr, ok := url.(string)
			Expect(ok).To(Equal(true))
			user1ShortenUrl = urlStr
		})

		It("should reject as url field is empty", func() {
			payload := fmt.Sprintf(`
			{
				"url": ""
			}
			`)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject due to invalid url", func() {
			payload := fmt.Sprintf(`
			{
				"url": "%v"
			}
			`, user1InvalidUrl)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject due to authorized problem", func() {
			payload := fmt.Sprintf(`
			{
				"url": "https://facebook.com"
			}
			`)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("Resolve a shorten url", func() {
		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/shortener/r/%v", user1ShortenUrl), nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusTemporaryRedirect))
			redirectUrl := recorder.Header().Get("Location")
			Expect(redirectUrl).To(Equal(user1Url))
		})

		It("should reject due to invalid request", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/shortener/r/12345678", nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Get user's urls", func() {
		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/url/list", nil)
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
			var resp shortener.URLsResponse
			err := getJSON(recorder.Result(), &resp)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(5 * time.Second)
			for _, url := range resp.URLs {
				if url.OriginURL == user1Url {
					Expect(url.Hits).To(Equal(int64(1))) // 1 hits by previous testing
				}
			}
		})

		It("should reject due to authorized problem", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/url/list", nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("Delete user's shorten url", func() {
		It("should reject due to authorized problem", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/user/url/r/%v", user1ShortenUrl), nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/user/url/r/%v", user1ShortenUrl), nil)
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))

			recorder = httptest.NewRecorder()
			req = httptest.NewRequest("GET", fmt.Sprintf("/api/shortener/r/%v", user1ShortenUrl), nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})

		It("should reject as the deletion request of shorten url doesn't exist", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", "/api/shortener/r/12345678", nil)
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Authenticate user", func() {
		It("should reject due to authorized problem", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/authCheck", nil)
			req.Header.Set("Cookie", "accessToken=123456")
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/authCheck", nil)
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

})

func getJSON(response *http.Response, target interface{}) error {
	return json.NewDecoder(response.Body).Decode(target)
}
