package main

import (
	"bytes"
	"context"
	"culturyus/models"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/jung-kurt/gofpdf"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/argon2"
)

type Template struct {
	templates *template.Template
}
type Cart struct {
	ProductName string
	Quantity    int
	Price       int
	PImg        string
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type dummyResponseWriter struct {
	header map[string][]string
}

func (d *dummyResponseWriter) Header() http.Header {
	if d.header == nil {
		d.header = make(map[string][]string)
	}
	return d.header
}

func (d *dummyResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (d *dummyResponseWriter) WriteHeader(statusCode int) {
}

var store = sessions.NewCookieStore([]byte("secret-key"))

func SetSession(r *http.Request, w http.ResponseWriter, key string, value interface{}) {
	session, _ := store.Get(r, "session-name")
	if existingVal, ok := session.Values[key].([]interface{}); ok {
		session.Values[key] = append(existingVal, value)
	} else {
		session.Values[key] = []interface{}{value}
	}
	session.Save(r, w)
}

func GetSession(r *http.Request, key string) interface{} {
	session, _ := store.Get(r, "session-name")
	return session.Values[key]
}

func encrypt(para string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(para), salt, 1, 64*1024, 4, 32)
	return base64.StdEncoding.EncodeToString(append(salt, hash...))
}

func compareHashes(hash1, hash2 []byte) bool {
	if len(hash1) != len(hash2) {
		return false
	}
	for i := range hash1 {
		if hash1[i] != hash2[i] {
			return false
		}
	}
	return true
}

func comparePassword(storedHash, password string) (bool, error) {
	storedHashBytes, err := base64.StdEncoding.DecodeString(storedHash)
	if err != nil {
		return false, err
	}
	salt := storedHashBytes[:16]
	storedPasswordHash := storedHashBytes[16:]
	computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	if !compareHashes(storedPasswordHash, computedHash) {
		return false, nil
	}

	return true, nil
}

func SendEmailAlertsForMultipleReqRejections(emails []string) (bool, error) {
	from := abc@gmail.com"
	password := "null"
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	subject := "Guide Request Update - No Response from Guide"
	auth := smtp.PlainAuth("", from, password, smtpHost)
	anySuccess := false
	for _, uemail := range emails {
		data, err := models.GetUserByEmail(uemail)
		if err != nil {
			log.Printf("User Data Fetch Failed for: %s, Error: %v\n", uemail, err)
			continue
		}
		to := []string{uemail}
		body := fmt.Sprintf(
			"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n"+
				"Dear %s,\r\n\r\n"+
				"We hope you're doing well!\r\n\r\n"+
				"It has been over 10 hours since you placed a guide request, and unfortunately, the guide hasn't responded yet.\r\n\r\n"+
				"We sincerely apologize for the inconvenience caused.\r\n"+
				"To avoid further delay in your travel plans, we kindly suggest checking for another available guide on CulturyUs.\r\n\r\n"+
				"Thank you for your understanding and patience.\r\n\r\n"+
				"Best Regards,\r\n"+
				"Team CulturyUs",
			from, uemail, subject, data.Fullname,
		)
		err = smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, from, to, []byte(body))
		if err != nil {
			log.Printf("Failed to send email to: %s, Error: %v\n", uemail, err)
			continue
		}
		_, err = models.RejectOldPendingRequests(uemail, 5)
		if err != nil {
			log.Printf("Failed to reject request for: %s, Error: %v\n", uemail, err)
			continue
		}
		anySuccess = true
	}
	if !anySuccess {
		return false, errors.New("failed to send email & reject request for all users")
	}
	return true, nil
}

func StartAutoRejectPendingRequests() {
	for {
		udata, mybug := models.GetPendingRequestUserEmails()
		if mybug != nil {
			fmt.Println("\nError Fetching Pending User Emails! GT kindly check whats wrong with system.\n")
		} else {
			if len(udata) == 0 {
				log.Println("No pending requests found to process.")
			} else {
				resultChan := make(chan bool)
				go func() {
					res, err := SendEmailAlertsForMultipleReqRejections(udata)
					if err != nil {
						log.Printf("Error occurred while sending emails: %v\n", err)
						resultChan <- false
						return
					}
					resultChan <- res
				}()

				go func() {
					myres := <-resultChan
					if myres {
						fmt.Println("\nEverything is great GT! CulturyUs on its Rocking way! Scheduled emails are sent successfully.\n")
					} else {
						fmt.Println("\nSomething went wrong GT, Kindly check!\n")
					}
				}()
			}
		}
		time.Sleep(30 * time.Minute)
	}
}

type CartItem struct {
	PName  string `json:"pname" bson:"pname"`
	PPrice int    `json:"pprice" bson:"pprice"`
	PQty   int    `json:"pqty" bson:"pqty"`
	PImg   string `json:"pimg" bson:"pimg"`
}

func getCart(c echo.Context) []CartItem {
	cookie, err := c.Cookie("cart")
	if err != nil {
		return []CartItem{}
	}

	decodedValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return []CartItem{}
	}

	var cart []CartItem
	err = json.Unmarshal([]byte(decodedValue), &cart)
	if err != nil {
		return []CartItem{}
	}

	return cart
}

func saveCart(c echo.Context, cart []CartItem) {
	data, err := json.Marshal(cart)
	if err != nil {
		return
	}

	encodedValue := url.QueryEscape(string(data))

	cookie := &http.Cookie{
		Name:     "cart",
		Value:    encodedValue,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	}
	c.SetCookie(cookie)
}
func convertCartItemsToModelsCart(cartItems []CartItem) []models.Cart {
	carts := make([]models.Cart, len(cartItems))
	for i, item := range cartItems {
		carts[i] = models.Cart{
			ProductName: item.PName,
			Quantity:    item.PQty,
			Price:       item.PPrice,
			PImg:        item.PImg,
		}
	}
	return carts
}

func main() {
	godotenv.Load()
	apiKey := os.Getenv("API_KEY")
	err := godotenv.Load("./.env")
	if err != nil {
		fmt.Println("Failed to load .env:", err)
	}

	if apiKey == "" {
		fmt.Println("API Key is missing!")
		return
	}

	models.Connect()
	gob.Register([]Cart{})
	dw := &dummyResponseWriter{}
	r, _ := http.NewRequest("GET", "/", nil)
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e.Renderer = t
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	go StartAutoRejectPendingRequests()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.POST("/register", func(c echo.Context) error {
		SetSession(r, dw, "fullname", c.FormValue("tb1"))
		SetSession(r, dw, "email", c.FormValue("tb2"))
		SetSession(r, dw, "dob", c.FormValue("tb3"))
		SetSession(r, dw, "addr", c.FormValue("tb4"))
		SetSession(r, dw, "contactno", c.FormValue("tb5"))
		SetSession(r, dw, "pwd", encrypt(c.FormValue("tb6")))
		sessions := GetSession(r, "pwd")
		sl := sessions.([]interface{})
		fmt.Println("Password is: ", sl[0])
		rand.Seed(time.Now().UnixNano())
		otp := rand.Intn(900000) + 100000
		fmt.Println("OTP is :", otp)
		from := "mypyschbuddy@gmail.com"
		pwd := "aoclddetchfgkscg"
		to := []string{c.FormValue("tb2")}
		smtphost := "smtp.gmail.com"
		port := "587"
		messagebody := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: CulturyUs OTP Verification Mail\r\n\r\nYour One Time Password for CulturyUs Verification is: %d", from, to[0], otp)
		msg := []byte(messagebody)
		auth := smtp.PlainAuth("", from, pwd, smtphost)
		err := smtp.SendMail(smtphost+":"+port, auth, from, to, msg)
		if err != nil {
			fmt.Printf("Failed to send OTP %v", err)
		}
		fmt.Println("OTP Sent.")
		SetSession(r, dw, "otp", otp)
		uemail := GetSession(r, "email")
		femail := uemail.([]interface{})
		vemail := fmt.Sprintf("%v", femail[0])
		dt := fmt.Sprintf("%v", time.Now())
		otpv := models.Otp_Verify{
			Email: vemail,
			Otp:   otp,
			DT:    dt,
		}
		res, err := models.InsertOTP_Verify(otpv)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("OTP Stored in DB", res)
		data := map[string]interface{}{
			"fullname": c.FormValue("tb1"),
			"contact":  c.FormValue("tb5"),
			"email":    c.FormValue("tb2"),
		}

		return c.Render(http.StatusOK, "otp.html", data)
	})

	e.POST("/response", func(c echo.Context) error {
		s1 := GetSession(r, "fullname")
		s2 := GetSession(r, "email")
		s3 := GetSession(r, "dob")
		s4 := GetSession(r, "addr")
		s5 := GetSession(r, "contactno")
		s6 := GetSession(r, "pwd")
		s7 := GetSession(r, "otp")
		ss1 := s1.([]interface{})
		ss2 := s2.([]interface{})
		ss3 := s3.([]interface{})
		ss4 := s4.([]interface{})
		ss5 := s5.([]interface{})
		ss6 := s6.([]interface{})
		ss7 := s7.([]interface{})
		fullname := fmt.Sprintf("%v", ss1[0])
		email := fmt.Sprintf("%v", ss2[0])
		dob := fmt.Sprintf("%v", ss3[0])
		addr := fmt.Sprintf("%v", ss4[0])
		contactno := fmt.Sprintf("%v", ss5[0])
		pwd := fmt.Sprintf("%v", ss6[0])
		otp := ss7[0]
		res2, err := models.GetOTPByEmail(email)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Email Got: ", res2)
		myotp, err := strconv.Atoi(c.FormValue("tb"))
		if myotp == otp {
			data := models.Users{
				Fullname:   fullname,
				Email:      email,
				DOB:        dob,
				Addr:       addr,
				ContactNo:  contactno,
				PWD:        pwd,
				Profilepic: "",
				Bio:        "",
				Status:     false,
				SkillCoins: 10,
			}
			res, err := models.InsertUsers(data)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Signup Successfull, go for signin...", res)
		} else {
			msg := "<html><script>alert('Validatio Failed! OTP does not match!')"
			return c.HTML(http.StatusOK, msg)
		}
		return c.Render(http.StatusOK, "response.html", nil)
	})

	e.POST("/signin", func(c echo.Context) error {
		username := strings.TrimSpace(c.FormValue("tb1"))
		pwd := strings.TrimSpace(c.FormValue("tb2"))
		//fmt.Println("Type of Pass: ", reflect.TypeOf(pwd))
		data, err := models.GetUserByEmail(username)
		if err != nil {
			fmt.Println("Error fetching user:", err)
			msg := "<html><script>alert('User not found!'); window.location='/';</script></html>"
			return c.HTML(http.StatusUnauthorized, msg)
		}
		dbusername := data.Email
		dbpass := data.PWD
		//fmt.Println(reflect.TypeOf(dbusername))
		fmt.Println(dbusername, dbpass)
		match, err := comparePassword(dbpass, pwd)
		var match1 bool = false
		if username == dbusername {
			match1 = true
		} else {
			match1 = false
		}
		if err != nil {
			log.Fatal(err)
		}
		if match1 && match {
			fmt.Println("Login successful!")
			res1, err1 := models.UpdateLoggedInUserStatus(username)
			res2, err2 := models.UpdateRestLoginStatus(username)
			if err1 != nil && err2 != nil {
				fmt.Println("Error1 -> ", err1, "\nError2 -> ", err2)
			}
			if res1 && res2 {
				return c.Render(http.StatusOK, "homepage.html", nil)
			} else {
				fmt.Println("Failed to login!")
			}
		} else {
			fmt.Println("Password mismatch")
			msg := "<html><script>alert('Invalid Username or Password !'); window.location='/';</script></html>"
			return c.HTML(http.StatusUnauthorized, msg)
		}

		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.POST("/home", func(c echo.Context) error {
		return c.Render(http.StatusAccepted, "homepage.html", nil)
	})

	e.GET("/skills", func(c echo.Context) error {
		data, err := models.RetrieveSkillData("Others")
		if err != nil {
			fmt.Println("Error -> ", err)
		}
		return c.Render(http.StatusOK, "skills.html", data)
	})

	e.GET("/profile", func(c echo.Context) error {
		data, err := models.GetCurrentUserInfo()
		if err != nil {
			log.Println("Error fetching user info:", err)
			return c.Render(http.StatusOK, "userProfile.html", map[string]interface{}{
				"Error": "Failed to retrieve user data",
			})
		}

		data2, err2 := models.RetrieveGuideData(data.Email)
		if err2 != nil {
			log.Println("No guide data found, using default values.")
			data2 = models.Tourist_Guide{}
		}
		checkIsGuide, err0 := models.CheckIfGuide(data.Email)
		if err0 != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong. Please try again after some time.'); window.location='/';</script>")
		}
		mydata, cause := models.RetrieveAllPurchasedSkills(context.Background(), data.Email)
		if cause != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong. Please try again after some time.'); window.location='/';</script>")
		}
		if checkIsGuide {
			reqData, bug := models.GetGuideBookingRequestData(data.Email)
			if bug != nil {
				return c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong. Please try again after some time.'); window.location='/';</script>")
			}
			return c.Render(http.StatusOK, "userProfile.html", map[string]interface{}{
				"data":  data,
				"data2": data2,
				"data3": reqData,
				"data4": mydata,
			})
		}
		return c.Render(http.StatusOK, "userProfile.html", map[string]interface{}{
			"data":  data,
			"data2": data2,
			"data4": mydata,
		})
	})

	e.POST("/changeProfile", func(c echo.Context) error {
		fullname := fmt.Sprintf("%v", c.FormValue("tb1"))
		addr := fmt.Sprintf("%v", c.FormValue("tb2"))
		contact := fmt.Sprintf("%v", c.FormValue("tb3"))
		bio := fmt.Sprintf("%v", c.FormValue("tb4"))
		imgfile, err := c.FormFile("tb5")
		if err != nil {
			log.Fatal(err)
		}
		src, err := imgfile.Open()
		if err != nil {
			fmt.Println("Error -> ", err)
		}
		defer src.Close()
		dbpath := "/static/img/usersDP/" + imgfile.Filename
		dirPath := "/home/ethicalgt/Documents/culturyus/static/img/usersDP"
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		}
		path := filepath.Join(dirPath, filepath.Base(imgfile.Filename))
		dst, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()
		data, err := models.GetCurrentUserInfo()
		if err != nil {
			fmt.Println("Error->", err)
		}
		email := data.Email
		res, err := models.UpdateUserInfo(email, fullname, addr, contact, bio, dbpath)
		if err != nil {
			fmt.Println("Error->", err)
		} else {
			fmt.Println("Profile Updated .", res)
			if op, err := io.Copy(dst, src); err != nil {
				fmt.Println("Error->", err, op)
			}
			fmt.Println("File stored in secured storage.")
			msg := "<html><body><script>alert('Profile Updated.');window.location='/profile';</script></body></html>"
			return c.HTML(http.StatusContinue, msg)
		}
		return c.Render(http.StatusOK, "userProfile.html", nil)
	})

	e.GET("/contributeSkills", func(c echo.Context) error {
		return c.Render(http.StatusContinue, "contributeSkill.html", nil)
	})

	e.POST("/contributeSkillAction", func(c echo.Context) error {
		data, err := models.GetCurrentUserInfo()
		if err != nil {
			log.Fatal(err)
		}
		email := fmt.Sprintf("%v", data.Email)
		uploader := fmt.Sprintf("%v", data.Fullname)
		skillname := fmt.Sprintf("%v", c.FormValue("tb1"))
		skilltype := fmt.Sprintf("%v", c.FormValue("tb3"))
		skilldesc := fmt.Sprintf("%v", c.FormValue("tb2"))
		mediatype := fmt.Sprintf("%v", c.FormValue("tb4"))
		instruct_lang := fmt.Sprintf("%v", c.FormValue("tb5"))
		datetime := fmt.Sprintf("%v", time.Now())
		file, err := c.FormFile("tb6")
		if err != nil {
			log.Fatal(err)
		}
		src, err := file.Open()
		if err != nil {
			fmt.Println("Error -> ", err)
		}
		defer src.Close()
		dbpath := "/static/skillResources/" + file.Filename
		dirPath := "/home/ethicalgt/Documents/culturyus/static/skillResources"
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		}
		path := filepath.Join(dirPath, filepath.Base(file.Filename))
		dst, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()
		op := models.Skills{
			Email:            email,
			Uploader:         uploader,
			Skillname:        skillname,
			Skilltype:        skilltype,
			Skilldesc:        skilldesc,
			Mediatype:        mediatype,
			Mediapath:        dbpath,
			LanguageInstruct: instruct_lang,
			UploadDateTime:   datetime,
		}
		res, err := models.InsertSkills(op)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("Contributed Successfully!", res.InsertedID)
			if op, err := io.Copy(dst, src); err != nil {
				fmt.Println("Error->", err, op)
			}
			if res, err := models.UpdateSkillCoinsOnUpload(); err != nil {
				fmt.Println("\nError->", err, res)
			}
			//fmt.Println("\nResult -> ", res.InsertedID, "\n")
			msg := "<html><body><script>alert('Thanks for your Contribution ,Your Contribution has been recorded. 5 Skillcoins has been added to you wallet. Have a good day!'); window.location='/contributeSkills';</script></body></html>"
			return c.HTML(http.StatusAccepted, msg)
		}

		return c.Render(http.StatusOK, "contributeSkill.html", nil)
	})

	e.POST("/skillcoinvalidator", func(c echo.Context) error {
		skilltype := c.FormValue("skillname")

		container2, err := models.CheckSkillAvailablity(skilltype)
		if err != nil {
			log.Printf("Error checking skill availability: %v\n", err)
			return c.HTML(http.StatusInternalServerError, "<script>alert('Something went wrong! Please try again later.'); window.location='/skills';</script>")
		}

		if container2 == nil {
			return c.HTML(http.StatusOK, "<script>alert('Oops! Skill is not available right now. We’ll notify you whenever it becomes available!'); window.location='/skills';</script>")
		}

		SetSession(r, dw, "skillname", skilltype)

		container, err := models.GetCurrentUserInfo()
		if err != nil {
			log.Printf("Error fetching user info: %v\n", err)
			return c.HTML(http.StatusOK, "<script>alert('No active user found. Please log in.'); window.location='/skills';</script>")
		}

		collexists, err1 := models.CheckIfDataExists(skilltype)
		if err1 != nil {
			log.Printf("Error checking skill data: %v\n", err1)
		}

		if collexists {
			rs, err := models.RetrievePurchasedSkillData(skilltype)
			if err != nil {
				log.Printf("Error retrieving purchased skill data: %v\n", err)
			}

			if rs.SkillPurchased == "Yes" {
				redirectHTML := `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Redirecting...</title>
	</head>
	<body>
		<script>
			function sendPostRequest() {
				let form = document.createElement('form');
				form.method = 'POST';
				form.action = '/skillpage';
	
				let input2 = document.createElement('input');
				input2.type = 'hidden';
				input2.name = 'user';
				input2.value = 'Ganesh Telore';
	
				form.appendChild(input2);
				document.body.appendChild(form);
				form.submit();
			}
			sendPostRequest();
		</script>
	</body>
	</html>`
				return c.HTML(http.StatusAccepted, redirectHTML)
			}
		}

		data := map[string]interface{}{
			"name":       container.Fullname,
			"skillcoins": container.SkillCoins,
			"skilltype":  container2.Skilltype,
		}

		return c.Render(http.StatusOK, "skillCoinValidator.html", data)
	})

	e.POST("/purchaseSkill", func(c echo.Context) error {
		dt := GetSession(r, "skillname")
		dt1 := dt.([]interface{})
		myskill := fmt.Sprintf("%v", dt1[0])
		container, err := models.GetCurrentUserInfo()
		if err != nil {
			log.Printf("Error fetching user info: %v\n", err)
			return c.HTML(http.StatusOK, "<script>alert('No active user found. Please log in.'); window.location='/skills';</script>")
		}
		mydt := fmt.Sprintf("%v", c.FormValue("formStatus"))
		fmt.Println("\nForm Value is here ->", mydt)
		var skillcoins int32 = int32(container.SkillCoins)

		balance := skillcoins - 3
		fmt.Println("\n balanced skillcoins ->", balance)
		if balance >= 0 {
			res, err := models.UpdateSkillCoinsOnPurchase(int(balance))
			if err != nil {
				log.Printf("Error updating skill coins: %v\n", err)
				return c.HTML(http.StatusInternalServerError, "<script>alert('Error processing your purchase. Please try again later.'); window.location='/skills';</script>")
			}
			fmt.Println(res)

			rawdata := models.PurchasedSkills{
				Email:                container.Email,
				SkillName:            myskill,
				SkillPurchased:       "Yes",
				PurchaseDateTime:     time.Now().Format("2006-01-02 15:04:05"),
				SkillStatus:          "Learning",
				SkillCertificateID:   "",
				SkillCertificatePath: "",
			}

			myres, err3 := models.InsertPurchasedSkillsData(rawdata)
			if err3 != nil {
				log.Printf("Error inserting purchased skill data: %v\n", err3)
				return c.HTML(http.StatusInternalServerError, "<script>alert('Something went wrong. Please try again later!'); window.location='/skills';</script>")
			}

			fmt.Println("\nSkill Purchased Successfully.", myres.InsertedID)
			msg := `<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Redirecting...</title>
	</head>
	<body>
		<script>
		alert('Skill unlocked Successfully.');
			function sendPostRequest() {
				let form = document.createElement('form');
				form.method = 'POST';
				form.action = '/skillpage';
	
				let input2 = document.createElement('input');
				input2.type = 'hidden';
				input2.name = 'user';
				input2.value = 'Ganesh Telore';
	
				form.appendChild(input2);
				document.body.appendChild(form);
				form.submit();
			}
			sendPostRequest();
		</script>
	</body>
	</html>`
			return c.HTML(http.StatusAccepted, msg)
		} else {
			return c.HTML(http.StatusOK, "<script>alert('SkillCoins are not sufficient. Kindly top up from your profile page.');window.location='/profile';</script>")
		}
	})

	e.POST("/skillpage", func(c echo.Context) error {
		dt := GetSession(r, "skillname")
		dt1 := dt.([]interface{})
		myskill := fmt.Sprintf("%v", dt1[0])
		data, err2 := models.GetCurrentUserInfo()
		if err2 != nil {
			log.Fatal("Error -> ", err2)
		}
		mydata, err := models.RetrieveSkillData(myskill)
		if err != nil {
			log.Fatal("Error ->", err)
		}
		data2 := map[string]interface{}{
			"skill":    mydata,
			"userdata": data,
		}
		return c.Render(http.StatusOK, "skillPage.html", data2)
	})

	e.GET("/generateCertificate", func(c echo.Context) error {
		data, err := models.GetCurrentUserInfo()
		if err != nil {
			log.Fatal("\nError ->", err)
		}
		dtt := GetSession(r, "skillname")
		dtt1 := dtt.([]interface{})
		skill := fmt.Sprintf("%v", dtt1[0])
		pdf := gofpdf.New("L", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFillColor(255, 204, 153)
		pdf.Rect(0, 0, 297, 210, "F")
		pdf.SetDrawColor(255, 69, 0)
		pdf.SetLineWidth(1)
		pdf.Rect(5, 5, 287, 200, "D")
		logoPath := "./static/img/logo.png"
		pdf.Image(logoPath, 10, 10, 20, 0, false, "", 0, "")
		pdf.SetFont("Helvetica", "B", 13)
		pdf.SetXY(10, 30)
		pdf.CellFormat(200, 10, "CulturyUS", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "B", 22)
		pdf.SetXY(50, 45)
		pdf.CellFormat(200, 10, "Certificate of Completion", "", 1, "C", false, 0, "")
		pdf.SetFont("Helvetica", "", 14)
		pdf.SetXY(50, 55)
		pdf.CellFormat(200, 10, "This is to certify that", "", 1, "C", false, 0, "")
		recipientName := data.Fullname
		pdf.SetFont("Helvetica", "B", 18)
		pdf.SetXY(50, 70)
		pdf.CellFormat(200, 10, recipientName, "", 1, "C", false, 0, "")
		courseName := skill
		pdf.SetFont("Helvetica", "", 14)
		pdf.SetXY(50, 85)
		pdf.CellFormat(200, 10, "has successfully completed the course on", "", 1, "C", false, 0, "")
		pdf.SetFont("Helvetica", "B", 16)
		pdf.SetXY(50, 100)
		pdf.CellFormat(200, 10, courseName, "", 1, "C", false, 0, "")
		currentDate := time.Now()
		Date := currentDate.Format("2 January 2006")
		completionDate := Date
		pdf.SetFont("Helvetica", "", 12)
		pdf.SetXY(50, 120)
		pdf.CellFormat(200, 10, "Date of Completion: "+completionDate, "", 1, "C", false, 0, "")
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetXY(50, 135)
		pdf.MultiCell(200, 6, "*This is NOT a professional certification. CulturyUS is not a verified institution. We are not responsible for misuse.*", "", "C", false)
		rand.Seed(time.Now().UnixNano())
		certificateID := fmt.Sprintf("CULTUS-CERT-%d-%d", rand.Intn(10000), rand.Intn(10000))
		pdf.SetFont("Helvetica", "", 14)
		pdf.SetXY(250, 177)
		pdf.CellFormat(0, 10, "Certificate ID: "+certificateID, "", 1, "R", false, 0, "")
		founderName := "Ganesh Telore"
		founderTitle := "Founder, CulturyUS"
		pdf.SetXY(10, 150)
		pdf.SetFont("Helvetica", "B", 14)
		pdf.CellFormat(200, 10, "Authorized By", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 14)
		pdf.CellFormat(200, 10, founderName, "", 1, "L", false, 0, "")
		pdf.CellFormat(200, 10, founderTitle, "", 1, "L", false, 0, "")
		signaturePath := "./static/img/signature.png"
		pdf.Image(signaturePath, 2, 180, 50, 0, false, "", 0, "")
		certPath := "./static/Certificates/" + skill + " CulturyUS_Certificate.pdf"

		err = pdf.OutputFileAndClose(certPath)
		if err != nil {
			panic(err)
		}

		dt, err2 := models.UpdateGeneratedCertificateInfo(skill, certificateID, certPath)
		if err2 != nil {
			log.Fatal("\nError ->", err2)
		}
		if dt {
			fmt.Println("\nCertificate Generated Successfully.")
		}
		res, err3 := models.RetrievePurchasedSkillData(skill)
		if err3 != nil {
			log.Fatal("\nError ->", err3)
		}
		fmt.Println("\nSkill ->", res.SkillCertificatePath)
		return c.Render(http.StatusOK, "certificatePage.html", res)
	})

	e.GET("/scpurchase", func(c echo.Context) error {
		msg := `<html>
    <body onload="submitForm();">
        <form id="paymentForm" method="POST" action="https://troubled-eloisa-ethicalpay-eb02efa7.koyeb.app/verifycred">
            <input type="hidden" name="tb1" value="EPOXtcMDl7">
            <input type="hidden" name="tb2" value="ZRRNCaAcNlbRRMOXdMhlevH79Z8iarJU">
            <input type="hidden" name="tb3" id="tb3" value="">
            <input type="hidden" name="tb4" value="http://localhost:2004/scpurchaseres">
        </form>

        <script>
            function submitForm() {
                let amount = prompt("Enter the amount to top up:");
                if (amount && !isNaN(amount) && amount > 0) {
                    document.getElementById("tb3").value = amount;
                    document.getElementById("paymentForm").submit();
                } else {
                    alert("Invalid amount. Please try again.");
                }
            }
        </script>
    </body>
</html>`
		return c.HTML(http.StatusContinue, msg)
	})
	e.POST("/scpurchaseres", func(c echo.Context) error {
		if c.Request().Method == "POST" {
			if c.FormValue("result") == "success" {
				mydata, myerr := models.GetCurrentUserInfo()
				if myerr != nil {
					log.Fatal(myerr)
				}
				data := models.Skillcoins_Transactions{
					Email:          mydata.Email,
					Payment_Status: c.FormValue("result"),
					Payment_Mode:   c.FormValue("paymode"),
					Amount:         c.FormValue("cost"),
					DateTime:       c.FormValue("datetime"),
				}
				res, err := models.InsertSkillcoinTransactionRec(data)
				if err != nil {
					fmt.Println(res)
					return c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong, Kindly try again after some time.');</script>")
				}
				cost, costerr := strconv.Atoi(c.FormValue("cost"))
				if costerr != nil {
					return c.HTML(http.StatusBadRequest, "<script>alert('Please enter valid money to top-up.');</script>")
				}
				if cost > 0 {
					updatedCoins := int(cost / 20)
					res2, err2 := models.UpdateSkillCoinsOnPayment(updatedCoins)
					if err2 != nil {
						return c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong, Skillcoins updation in wallet failed! kindly wait for some time.');</script>")
					}
					fmt.Println(res2)
				}
			}
		}
		status := c.FormValue("result")
		return c.Render(http.StatusAccepted, "scresponsePage.html", map[string]interface{}{
			"paymentstatus": c.FormValue("result"),
			"successed":     status == "success",
			"paymentmode":   c.FormValue("paymode"),
			"amount":        c.FormValue("cost"),
			"datetime":      c.FormValue("datetime"),
		})
	})

	e.GET("/tourguide", func(c echo.Context) error {
		return c.Render(http.StatusOK, "touristGuide.html", nil)
	})

	e.POST("/registerGuide", func(c echo.Context) error {
		mydata, err1 := models.GetCurrentUserInfo()
		if err1 != nil {
			return c.HTML(http.StatusOK, "<script>alert('No Active User Found'); window.location='/';</script>")
		}
		docfile, err := c.FormFile("tb5")
		if err != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('File upload failed!'); window.location='/profile';</script>")
		}
		src, err := docfile.Open()
		if err != nil {
			fmt.Println("Error opening file:", err)
			return c.HTML(http.StatusBadRequest, "<script>alert('Error processing file!'); window.location='/profile';</script>")
		}
		defer src.Close()
		dirPath := "/home/ethicalgt/Documents/culturyus/static/guideDoc"
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		path := filepath.Join(dirPath, filepath.Base(docfile.Filename))
		dst, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()
		if _, err := io.Copy(dst, src); err != nil {
			return fmt.Errorf("failed to copy file content: %v", err)
		}
		err = os.Chmod(path, 0644)
		if err != nil {
			fmt.Println("Failed to set permissions:", err)
		}
		fmt.Println("File saved at:", path)
		dbpath := "/static/guideDoc/" + docfile.Filename
		data := models.Tourist_Guide{
			Fullname:         mydata.Fullname,
			Email:            mydata.Email,
			DOB:              mydata.DOB,
			Addr:             mydata.Addr,
			ContactNo:        mydata.ContactNo,
			Profilepic:       mydata.Profilepic,
			Experience:       c.FormValue("tb1"),
			Languages:        c.FormValue("tb2"),
			Preffered_States: c.FormValue("tb3"),
			Charges:          c.FormValue("tb4"),
			Identity_Proof:   dbpath,
			Bio:              c.FormValue("tb6"),
			Active:           true,
		}
		res, err2 := models.InsertGuide(data)
		if err2 != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong!'); window.location='/profile';</script>")
		}
		fmt.Println(res)
		fmt.Println("\nDocument Saved Securely.")
		return c.HTML(http.StatusOK, `
    <html>
    <head><script>
        alert('Congratulations! You\'re now a part of CulturyUs Guides!');
        window.location='/profile';
    </script></head>
    <body></body>
    </html>
`)

	})

	e.POST("/bookGuide", func(c echo.Context) error {
		mydata, err := models.GetCurrentUserInfo()
		if err != nil {
			return c.HTML(http.StatusOK, "<script>alert('No Active User Found!'); window.location='/';</script>")
		}
		fmt.Println(mydata.Email)
		data, err2 := models.RetrieveAllGuides(c.FormValue("state"))
		if err2 != nil {
			return c.HTML(http.StatusOK, "<script>alert('Something went wrong, please try again later.'); window.location='/tourguide';</script>")
		}
		if len(data) == 0 {
			return c.HTML(http.StatusAccepted, "<script>alert('Oops! No guides available for the selected state yet!'); window.location='/tourguide';</script>")
		}
		return c.Render(http.StatusOK, "touristGuidesPage.html", map[string]interface{}{
			"data": data,
		})
	})

	e.POST("/requestBookingGuide", func(c echo.Context) error {

		from := "mypyschbuddy@gmail.com"
		pwd := "aoclddetchfgkscg"
		smtphost := "smtp.gmail.com"
		port := "587"

		guideEmail := c.FormValue("guideemail")
		if guideEmail == "" {
			return c.HTML(http.StatusBadRequest, "<script>alert('Guide email is missing!'); window.location='/bookguide';</script>")
		}
		data, err2 := models.RetrieveGuideData(guideEmail)
		if err2 != nil {
			return c.HTML(http.StatusOK, "<script>alert('Something went wrong, please try again later.'); window.location='/tourguide';</script>")
		}
		to := []string{guideEmail}
		messagebody := fmt.Sprintf(
			"From: %s\r\nTo: %s\r\nSubject: Reminder: New Booking Request on CulturyUs\r\n\r\nDear %s,\r\n\r\nYou have a new booking request on CulturyUs! Please log in to your profile to review and confirm your availability.\r\n\r\nTimely responses help ensure a great experience for travelers.\r\n\r\nBest regards,\r\nCulturyUs Team",
			from, guideEmail, data.Fullname,
		)
		msg := []byte(messagebody)
		auth := smtp.PlainAuth("", from, pwd, smtphost)
		addr := fmt.Sprintf("%s:%s", smtphost, port)

		err := smtp.SendMail(addr, auth, from, to, msg)
		if err != nil {
			fmt.Printf("Error Occured -> %v\n", err)
			return c.HTML(http.StatusOK, "<script>alert('Something went wrong! Please try again later.'); window.location='/tourguide';</script>")
		}
		userData, err := models.GetCurrentUserInfo()
		if err != nil {
			return c.HTML(http.StatusOK, "<script>alert('No Active User Found, Kindly Login'); window.location='/';</script>")
		}
		mydata := models.Guide_Requests{
			Datetime:   time.Now(),
			UserEmail:  userData.Email,
			GuideEmail: data.Email,
			Msg: `Hello!
I'm interested in booking a guide through CulturyUs. If you're available, we can discuss further details like the destination, city, time, and other specifics of the tour. I'm also ready to pay an advance of ₹100 for the booking when we meet at the designated location.

Looking forward to your response!`,
			Status:           "pending",
			Rejection_Reason: "",
		}
		res, err := models.InsertGuideRequestData(mydata)
		if err != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('Sorry only 1 request is allowed at one time, kindly try again after 10 hours!'); window.location='/tourguide';</script>")
		}
		fmt.Print("Guide request document updated successfully.", res)
		return c.HTML(http.StatusOK, "<script>alert('Booking request has been sent to Guide. Kindly wait for his/her response.');window.location='/tourguide';</script>")
	})

	e.GET("/GIA", func(c echo.Context) error {
		return c.Render(http.StatusOK, "GIA.html", nil)
	})

	e.POST("/bookingAccepted", func(c echo.Context) error {
		gemail := c.FormValue("gemail")
		uemail := c.FormValue("uemail")
		data, err := models.GetUserByEmail(gemail)
		if err != nil {
			c.HTML(http.StatusBadRequest, "<script>alert('Failed to sent rejection confirmation. please try again later.'); window.location='/profile';</script>")
		}
		data2, err2 := models.GetUserByEmail(uemail)
		if err2 != nil {
			c.HTML(http.StatusBadRequest, "<script>alert('Failed to sent rejection confirmation. please try again later.'); window.location='/profile';</script>")
		}
		from := "mypyschbuddy@gmail.com"
		password := "aoclddetchfgkscg"
		smtpHost := "smtp.gmail.com"
		smtpPort := "587"

		to := []string{uemail}

		subject := "Guide Booking Request Acceptance – CulturyUs"
		body := fmt.Sprintf(
			"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n"+
				"Dear %s,\r\n\r\n"+
				"Congratulations! Your request for a guide has been accepted by CulturyUs.\r\n\r\n"+
				"Here are the contact details of your guide:\r\n\r\n"+
				"Name     : %s\r\n"+
				"Phone    : %s\r\n"+
				"Email    : %s\r\n"+
				"Location : %s\r\n\r\n"+
				"Please feel free to contact the guide directly for further communication and planning.\r\n\r\n"+
				"Thank you for choosing CulturyUs. We wish you a wonderful and culturally enriching experience!\r\n\r\n"+
				"Warm regards,\r\nTeam CulturyUs",
			from, to, subject, data2.Fullname, data.Fullname, data.ContactNo, data.Email, data.Addr)

		auth := smtp.PlainAuth("", from, password, smtpHost)
		errr := smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, from, to, []byte(body))

		if errr != nil {
			fmt.Printf("Error Occurred -> %v\n", err)
			return c.HTML(http.StatusInternalServerError, "<script>alert('Something went wrong! Please try again later.'); window.location='/tourguide';</script>")
		}
		res, bug := models.AcceptApprovalRequest(uemail)
		if bug != nil {
			c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong please try again later!'); window.location='/profile';</script>")
		}
		if res {
			fmt.Println("\n Record Updated.")
		}
		return c.HTML(http.StatusOK, "<script>alert('Your Response has been recorded & forwarded to tourist. Thank You!'); window.location='/profile';</script>")
	})

	e.POST("/bookingRejected", func(c echo.Context) error {
		msg := c.FormValue("msg")
		uemail := c.FormValue("uemail")
		data, err := models.GetUserByEmail(uemail)
		if err != nil {
			c.HTML(http.StatusBadRequest, "<script>alert('Failed to sent rejection confirmation. please try again later.'); window.location='/profile';</script>")
		}
		from := "mypyschbuddy@gmail.com"
		password := "aoclddetchfgkscg"
		smtpHost := "smtp.gmail.com"
		smtpPort := "587"

		to := []string{uemail}

		subject := "Guide Booking Unavailable – CulturyUs"
		body := fmt.Sprintf(
			"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n"+
				"Dear %s,\r\n\r\n"+
				"Thank you for your interest in collaborating with CulturyUs.\r\n\r\n"+
				"We regret to inform you that your guide request has been declined at this time. We sincerely apologize for the inconvenience caused.\r\n\r\n"+
				"We truly appreciate your enthusiasm and hope to explore opportunities to work together in the future.\r\n"+
				"If you have any questions or would like to stay updated on upcoming openings, feel free to reach out to us.\r\n\r\n"+
				"The Reason: %s\r\n\r\n"+
				"Warm regards,\r\nTeam CulturyUs",
			from, to, subject, data.Fullname, msg)

		auth := smtp.PlainAuth("", from, password, smtpHost)
		errr := smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, from, to, []byte(body))

		if errr != nil {
			fmt.Printf("Error Occurred -> %v\n", err)
			return c.HTML(http.StatusInternalServerError, "<script>alert('Something went wrong! Please try again later.'); window.location='/tourguide';</script>")
		}
		res, bug := models.RejectApprovalRequest(uemail)
		if bug != nil {
			c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong please try again later!'); window.location='/profile';</script>")
		}
		if res {
			fmt.Println("\n Record Updated.")
		}
		res2, bug2 := models.UpdateRejectionReason(uemail, msg)
		if bug2 != nil {
			log.Println(bug2)
		}
		if res2 {
			fmt.Println("\nReason Updated!")
		}
		return c.HTML(http.StatusOK, "<script>alert('Your Response has been recorded & forwarded to tourist. Thank You!'); window.location='/profile';</script>")
	})

	e.POST("/chat", func(c echo.Context) error {
		var body struct {
			Prompt string `json:"prompt"`
		}
		if err := c.Bind(&body); err != nil || body.Prompt == "" {
			return c.String(http.StatusBadRequest, "Missing prompt")
		}
		reqBody := map[string]interface{}{
			"model": "mistralai/mistral-7b-instruct",
			"messages": []map[string]string{
				{"role": "user", "content": body.Prompt},
			},
		}
		jsonBody, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Request failed")
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		fmt.Println(raw)
		var result struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		json.Unmarshal(raw, &result)
		if len(result.Choices) > 0 {
			return c.JSON(http.StatusOK, map[string]string{
				"response": result.Choices[0].Message.Content,
			})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"response": "No response from model",
		})
	})

	e.GET("/news", func(c echo.Context) error {
		const apiKey = "2e4ddaaa9ef140168a0c75df8d2df870"

		type News struct {
			Title       string `json:"title"`
			Author      string `json:"author"`
			URLToImage  string `json:"urlToImage"`
			Description string `json:"description"`
			Content     string `json:"content"`
			PublishedAt string `json:"publishedAt"`
			Source      struct {
				Name string `json:"name"`
			} `json:"source"`
			URL string `json:"url"`
		}

		type NewsResponse struct {
			Status   string `json:"status"`
			Articles []News `json:"articles"`
		}

		url := fmt.Sprintf("https://newsapi.org/v2/everything?q=indian%%20culture&language=en&sortBy=publishedAt&pageSize=15&apiKey=%s", apiKey)

		resp, err := http.Get(url)
		if err != nil {
			log.Println("Error fetching news:", err)
			return c.Render(http.StatusInternalServerError, "news.html", map[string]interface{}{
				"message": "Failed to fetch news. Try again later.",
			})
		}
		defer resp.Body.Close()

		var newsResponse NewsResponse
		if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
			log.Println("Error decoding news response:", err)
			return c.Render(http.StatusInternalServerError, "news.html", map[string]interface{}{
				"message": "Failed to process news. Try again later.",
			})
		}

		if newsResponse.Status != "ok" || len(newsResponse.Articles) == 0 {
			return c.Render(http.StatusNotFound, "news.html", map[string]interface{}{
				"message": "No news found on Indian culture.",
			})
		}

		var output []map[string]interface{}
		for _, article := range newsResponse.Articles {
			if article.URLToImage == "" {
				article.URLToImage = "/static/img/news-img.jpg"
			}

			output = append(output, map[string]interface{}{
				"Title":       article.Title,
				"Author":      article.Author,
				"Source":      article.Source.Name,
				"PublishedAt": article.PublishedAt,
				"Content":     article.Content,
				"Description": article.Description,
				"ImageURL":    article.URLToImage,
				"ReadMoreURL": article.URL,
			})
		}

		return c.Render(http.StatusOK, "news.html", map[string]interface{}{
			"articles": output,
		})
	})

	e.GET("/forum", func(c echo.Context) error {
		messages, err := models.GetForumMessages(30)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to fetch messages")
		}
		data, err2 := models.GetCurrentUserInfo()
		if err2 != nil {
			return c.String(http.StatusInternalServerError, "Failed to fetch messages")
		}
		return c.Render(http.StatusOK, "forum.html", map[string]interface{}{
			"Messages": messages,
			"data":     data,
		})
	})
	e.POST("/forum/message", func(c echo.Context) error {
		data, err2 := models.GetCurrentUserInfo()
		if err2 != nil {
			return c.String(http.StatusInternalServerError, "Failed to fetch messages")
		}
		var msg models.Chat_forum
		if err := c.Bind(&msg); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
		}
		msg.Datetime = time.Now()
		msg.Email = data.Email
		msg.ProfilePic = data.Profilepic
		res, err := models.InsertChatInForum(msg)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to insert message"})
		}
		msg.ID = res.InsertedID.(primitive.ObjectID)
		return c.JSON(http.StatusOK, msg)
	})
	e.GET("/forum/messages", func(c echo.Context) error {
		messages, err := models.GetForumMessages(50)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch messages"})
		}
		return c.JSON(http.StatusOK, messages)
	})

	e.GET("/marketplace", func(c echo.Context) error {
		data, err := models.RetrieveArtpiecesData()
		if err != nil {
			fmt.Println("No artpieces uploaded yet.")
		}
		return c.Render(http.StatusOK, "marketplace.html", data)
	})

	e.POST("/sellartpieces", func(c echo.Context) error {
		return c.Render(http.StatusAccepted, "sellartpieces.html", nil)
	})

	e.POST("/uploadartpieces", func(c echo.Context) error {
		pname := c.FormValue("tb1")
		pdesc := c.FormValue("tb2")
		pprice := c.FormValue("tb3")
		imgfile, err := c.FormFile("tb4")
		if err != nil {
			log.Fatal(err)
		}
		src, err := imgfile.Open()
		if err != nil {
			fmt.Println("Error -> ", err)
		}
		defer src.Close()
		dbpath := "/static/img/artpieces_img/" + imgfile.Filename
		dirPath := "/home/ethicalgt/Documents/culturyus/static/img/artpieces_img"
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		}
		path := filepath.Join(dirPath, filepath.Base(imgfile.Filename))
		dst, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()
		currentuserdata, err := models.GetCurrentUserInfo()
		if err != nil {
			return c.HTML(http.StatusAccepted, "<script>alert('No Active user session found. Kindly login!); window.location='/';</script>")
		}
		data := models.Artpieces{
			Datetime:      time.Now(),
			Email:         currentuserdata.Email,
			Artpiece_name: pname,
			Artpiece_desc: pdesc,
			Artpiece_img:  dbpath,
			Price:         pprice,
		}
		res, err2 := models.InsertArtpeices(data)
		if err2 != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('Something went wrong with marketplace. Please try again later.'); window.location='/marketplace';</script>")
		}
		fmt.Println("Artpieces Updated in DB.", res)
		if op, err := io.Copy(dst, src); err != nil {
			fmt.Println("Error->", err, op)
		}
		fmt.Println("File stored in secured storage.")
		msg := `<html><body><script>alert('Artpiece registration successfull.'); 
		const form = document.createElement("form");
    form.method="POST";
    form.action="/sellartpieces";
    document.body.appendChild(form);
    form.submit();
	</script></body></html>`
		return c.HTML(http.StatusAccepted, msg)
	})

	e.POST("/addtocart", func(c echo.Context) error {
		item := CartItem{}

		item.PName = c.FormValue("pname")
		item.PImg = c.FormValue("pimg")

		priceStr := c.FormValue("pprice")
		qtyStr := c.FormValue("pquantity")

		price, err := strconv.Atoi(priceStr)
		if err != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('Invalid price.'); window.location='/marketplace';</script>")
		}
		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('Invalid quantity.'); window.location='/marketplace';</script>")
		}

		item.PPrice = price
		item.PQty = qty

		cart := getCart(c)
		cart = append(cart, item)
		saveCart(c, cart)

		return c.HTML(http.StatusOK, "<script>alert('Item added to cart.'); window.location='/marketplace';</script>")
	})

	e.GET("/cart", func(c echo.Context) error {
		data := getCart(c)
		fmt.Printf("\nCart data: %+v\n", data)
		return c.Render(http.StatusOK, "cart.html", map[string]interface{}{
			"CartItems": data,
		})

	})

	e.POST("/removefromcart", func(c echo.Context) error {
		pname := c.FormValue("pname")
		cart := getCart(c)
		newCart := []CartItem{}
		for _, item := range cart {
			if item.PName != pname {
				newCart = append(newCart, item)
			}
		}
		saveCart(c, newCart)
		return c.HTML(http.StatusOK, "<script>alert('Item removed'); window.location='/cart';</script>")
	})

	e.POST("/preorder", func(c echo.Context) error {
		cost := c.FormValue("cost")
		datetime := c.FormValue("datetime")
		paymode := c.FormValue("paymode")
		status := c.FormValue("result")

		msg := `<html><body><script>
			const form = document.createElement('form');
			form.method = 'POST';
			form.action = '/placeorder';
	
			const inputCost = document.createElement('input');
			inputCost.type = 'hidden';
			inputCost.name = 'cost';
			inputCost.value = '` + cost + `';
			form.appendChild(inputCost);
	
			const inputDatetime = document.createElement('input');
			inputDatetime.type = 'hidden';
			inputDatetime.name = 'datetime';
			inputDatetime.value = '` + datetime + `';
			form.appendChild(inputDatetime);
	
			const inputPaymode = document.createElement('input');
			inputPaymode.type = 'hidden';
			inputPaymode.name = 'paymode';
			inputPaymode.value = '` + paymode + `';
			form.appendChild(inputPaymode);
	
			const inputStatus = document.createElement('input');
			inputStatus.type = 'hidden';
			inputStatus.name = 'status';
			inputStatus.value = '` + status + `';
			form.appendChild(inputStatus);
	
			document.body.appendChild(form);
			form.submit();
		</script></body></html>`

		return c.HTML(http.StatusAccepted, msg)
	})

	e.POST("/placeorder", func(c echo.Context) error {
		cart := getCart(c)

		if len(cart) == 0 {
			return c.HTML(http.StatusBadRequest, "<script>alert('Cart is empty'); window.location='/marketplace';</script>")
		}
		udata, err := models.GetCurrentUserInfo()
		if err != nil {
			return c.HTML(http.StatusBadRequest, "<script>alert('No active user found, kindly login!'); window.location='/';<script>")
		}
		email := udata.Email
		paymode := c.FormValue("paymode")
		amountStr := c.FormValue("cost")

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return c.HTML(http.StatusInternalServerError, "<script>alert('Invalid amount retrieved. Kindly check the amount again!'); window.location='/cart';</script>")
		}

		modelCart := convertCartItemsToModelsCart(cart)
		fmt.Println("\nConverted cart:", modelCart)
		itemsSummary := models.ConvertCartToSummary(modelCart)
		fmt.Println("\nSummary:", itemsSummary)

		order := &models.Orders{
			Email:    email,
			Items:    itemsSummary,
			PayMode:  paymode,
			Amount:   amount,
			DateTime: primitive.NewDateTimeFromTime(time.Now()),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if c.FormValue("status") == "success" {
			_, err = models.InsertOrder(ctx, order)
			if err != nil {
				return c.HTML(http.StatusInternalServerError, "<script>alert('Failed to place order. Please try again later.'); window.location='/cart';</script>")
			}

			cookie := &http.Cookie{
				Name:     "cart",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			}
			c.SetCookie(cookie)
		}
		status := c.FormValue("status")
		fmt.Println("\n Payment Status: ", status)
		return c.Render(http.StatusOK, "orderAcknowTemplate.html", map[string]interface{}{
			"paymentstatus": c.FormValue("status"),
			"successed":     status == "success",
			"paymentmode":   c.FormValue("paymode"),
			"amount":        c.FormValue("cost"),
			"datetime":      c.FormValue("datetime"),
		})
	})

	e.Static("/static", "static")
	fmt.Println("CulturyUs Boomed at http://localhost:2004")
	e.Logger.Fatal(e.Start(":2004"))
	models.Disconnect()

}
