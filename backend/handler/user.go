package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"khu-capstone-18-backend/auth"
	"khu-capstone-18-backend/model"
	"khu-capstone-18-backend/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SIGNUP HANDLER START")
	req := model.User{}

	b, _ := io.ReadAll(r.Body)

	if err := json.Unmarshal(b, &req); err != nil {
		fmt.Println("UNMARSHAL ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// DB에 유저 삽입
	id := uuid.NewString()
	if err := repository.CreateUser(id, req.Username, req.Password, req.Email, req.Weight); err != nil {
		fmt.Println("CREATE USER ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// JWT 토큰 생성
	token, err := auth.GenerateJwtToken(id, time.Hour*1200)
	if err != nil {
		fmt.Println("GENERATE JWT TOKEN ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := struct {
		Message string `json:"message"`
		UserId  string `json:"user_id"`
		Token   string `json:"token"`
	}{
		Message: "Signup successful",
		Token:   token,
	}

	// 응답
	response, err := json.Marshal(resp)
	if err != nil {
		fmt.Println("JSON MARSHALING ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("LOGIN HANDLER START")
	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	b, _ := io.ReadAll(r.Body)

	if err := json.Unmarshal(b, &req); err != nil {
		fmt.Println("UNMARSHAL ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	pw, err := repository.GetPasswordByUsername(req.Username)
	if err != nil {
		fmt.Println("GET PASSWORD ERR:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if pw != req.Password {
		fmt.Println("USER " + req.Username + " TRIED TO LOGIN WITH UNCORRECT PASSWORD")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	uid, err := repository.GetUserID(req.Username)
	if err != nil {
		fmt.Println("GETTING USER ID ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// JWT 생성
	token, err := auth.GenerateJwtToken(uid, time.Hour*1200)
	if err != nil {
		fmt.Println("GENERATE JWT ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 응답
	response := struct {
		Message string `json:"message"`
		Token   string `json:"token"`
		UserID  string `json:"user_id"`
	}{
		Message: "Login successful",
		Token:   token,
		UserID:  uid,
	}

	resp, err := json.Marshal(response)
	if err != nil {
		fmt.Println("JSON MARSHALING ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		fmt.Println("NO JWT TOKEN EXIST ERROR")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Bearer 토큰 추출
	t := authHeader[7:]

	username, err := auth.ValidateJwtToken(t)
	if err != nil {
		fmt.Println("JWT TOKEN VALIDATION ERR:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 토큰 삭제
	if _, err := auth.GenerateJwtToken(username, 0); err != nil {
		fmt.Println("JWT TOKEN REMOVE ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 클라이언트에게 만료된 토큰 반환
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, "Logout successful")))
}

func OptionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	w.WriteHeader(http.StatusOK)
	return
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		fmt.Println("NO JWT TOKEN EXIST ERROR")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Bearer 토큰 추출
	t := authHeader[7:]

	requestor, err := auth.ValidateJwtToken(t)
	if err != nil {
		fmt.Println("JWT TOKEN VALIDATION ERR:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u, err := repository.GetUser(requestor)
	if err != nil {
		fmt.Println("GET USER PROFILE ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bestDistance, bestTime, err := GetBestRecordByUserId(userId)
	if err != nil {
		fmt.Println("USER DOES NOT HAVE BEST RECORD")
		bestDistance = "0.00"
		bestTime = 0.0
	}
	f, _ := strconv.ParseFloat(bestDistance, 64)

	totalDistance, totalTime, err := repository.GetUserTotalRecord(userId)
	if err != nil {
		fmt.Println("GET USER'S TOTAL RECORD DATA ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 응답
	response := struct {
		UserID        string  `json:"user_id"`
		Username      string  `json:"username"`
		ProfileImage  string  `json:"profile_image"`
		TotalDistance float64 `json:"total_distance"`
		TotalTime     int     `json:"total_time"`
		BestRecord    struct {
			Distance float64 `json:"distance"`
			Time     int     `json:"time"`
		} `json:"best_record"`
		WeeklyGoal string `json:"weekly_goal"`
		Nickname   string `json:"nickname"`
	}{
		Username:      u.Username,
		ProfileImage:  u.ProfileImage,
		TotalDistance: totalDistance,
		TotalTime:     int(totalTime.Seconds()),
		BestRecord: struct {
			Distance float64 "json:\"distance\""
			Time     int     "json:\"time\""
		}{
			Distance: f,
			Time:     int(bestTime.Seconds()),
		},
		WeeklyGoal: u.WeeklyGoal,
		Nickname:   u.Nickname,
	}

	resp, err := json.Marshal(response)
	if err != nil {
		fmt.Println("JSON MARSHALING ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		fmt.Println("NO JWT TOKEN EXIST ERROR")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Bearer 토큰 추출
	t := authHeader[7:]

	userId, err := auth.ValidateJwtToken(t)
	if err != nil {
		fmt.Println("JWT TOKEN VALIDATION ERR:", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	b, _ := io.ReadAll(r.Body)
	req := repository.User{}
	if err := json.Unmarshal(b, &req); err != nil {
		fmt.Println("UNMARSHAL ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := repository.PutUser(userId, req.ProfileImage, req.WeeklyGoal); err != nil {
		fmt.Println("PUT USER PROFILE ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 응답
	response := struct {
		Message string `json:"message"`
	}{
		Message: "Profile updated successfully.",
	}

	resp, err := json.Marshal(response)
	if err != nil {
		fmt.Println("JSON MARSHALING ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func GetBestRecordByUserId(userId string) (bestDistance string, bestTime time.Duration, err error) {
	return repository.GetUserBestRecord(userId)
}

func GetTotalDistanceAndTime(userId string) (totalDistance float64, totalTime time.Duration, err error) {
	records, err := repository.GetTotalSessions(userId)
	if err != nil {
		return 0, 0, err
	}

	d := 0.0
	var t time.Duration

	for _, r := range *records {
		d += r.Distance
		tmp, _ := time.ParseDuration(r.Time)
		t += tmp
	}

	return d, t, nil
}
