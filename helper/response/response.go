package response

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"dario.cat/mergo"
	"github.com/labstack/echo/v4"
)

type ResponseHelper struct {
	SetupGitInfo bool
	BranchName   string
	Hash         string
	Updated      string
	Hostname     string
}

type responseFormat struct {
	C       echo.Context
	Code    int
	Status  string
	Message string
	Data    interface{}
	Field   interface{}
}

func (r *ResponseHelper) SetResponse(c echo.Context, code int, status string, message string, data interface{}, field interface{}) responseFormat {
	return responseFormat{c, code, status, message, data, field}
}

func (r *ResponseHelper) SendResponse(res responseFormat) error {
	if len(res.Message) == 0 {
		res.Message = http.StatusText(res.Code)
	}
	r.LoadGitInfo()
	response := map[string]interface{}{
		"git_branch":  r.BranchName,
		"git_hash":    r.Hash,
		"git_updated": r.Updated,
		"hostname":    r.Hostname,
		"code":        res.Code,
		"status":      res.Status,
		"message":     res.Message,
		"data":        res.Data,
		"field":       res.Field,
	}

	if res.Field != nil {
		mergo.Merge(&response, map[string]interface{}{
			"field": res.Field,
		})
	}

	return res.C.JSON(res.Code, response)
}

func (r *ResponseHelper) GetBranch() string {
	branchName := ""

	firstLineHead := ""
	file, err := os.Open(".git/HEAD")
	if err != nil {
		// log.Fatal(err)
		// return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		firstLineHead = scanner.Text()
		break
	}
	headInfo := strings.Split(firstLineHead, "/")
	if len(headInfo) >= 2 {
		branchName = headInfo[2]
	}

	return branchName
}

func (r *ResponseHelper) GetHash(branchName string) string {
	hash := ""
	log.Println(branchName)
	file, err := os.Open(".git/ORIG_HEAD")
	if err != nil {
		// log.Fatal(err)
		// return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hash = scanner.Text()
		break
	}

	return hash
}

func (r *ResponseHelper) GetUpdated() string {
	statinfo, err := os.Stat(".git/index")
	if err != nil {
		// log.Fatal(err)
		// return ""
	}

	return statinfo.ModTime().Format("2006-01-02 15:04:05")
}

func (r *ResponseHelper) GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "-"
	}

	return hostname
}

func (r *ResponseHelper) LoadGitInfo() {
	if r.SetupGitInfo {
		return
	}
	r.BranchName = r.GetBranch()
	r.Hash = r.GetHash(r.BranchName)
	r.Updated = r.GetUpdated()
	r.Hostname = r.GetHostname()
	r.SetupGitInfo = true
}

func (r *ResponseHelper) SendSuccess(c echo.Context, message string, data interface{}) error {
	res := r.SetResponse(c, http.StatusOK, "success", message, data, nil)
	return r.SendResponse(res)
}

func (r *ResponseHelper) SendBadRequest(c echo.Context, message string, data interface{}) error {
	res := r.SetResponse(c, http.StatusBadRequest, "error", message, data, nil)
	return r.SendResponse(res)
}

func (r *ResponseHelper) SendError(c echo.Context, message string, data interface{}) error {
	res := r.SetResponse(c, http.StatusInternalServerError, "error", message, data, nil)
	return r.SendResponse(res)
}

func (r *ResponseHelper) SendUnauthorized(c echo.Context, message string, data interface{}) error {
	res := r.SetResponse(c, http.StatusUnauthorized, "error", message, data, nil)
	return r.SendResponse(res)
}

func (r *ResponseHelper) SendForbidden(c echo.Context, message string, data interface{}) error {
	res := r.SetResponse(c, http.StatusForbidden, "error", message, data, nil)
	return r.SendResponse(res)
}

func (r *ResponseHelper) SendNotFound(c echo.Context, message string, data interface{}) error {
	res := r.SetResponse(c, http.StatusNotFound, http.StatusText(http.StatusNotFound), message, data, nil)
	return r.SendResponse(res)
}

type ResponsePagination struct {
	Records            interface{} `json:"records"`
	TotalRecord        int64       `json:"total_record"`
	TotalRecordPerPage int64       `json:"total_record_per_page"`
	TotalRecordSearch  int64       `json:"total_record_search"`
	TotalPage          int64       `json:"total_page"`
	CurrentPage        int         `json:"current_page"`
	RowNumberStart     int         `json:"row_number_start"`
	RowNumberEnd       int         `json:"row_number_end"`
}

func (r *ResponseHelper) SendPaginationResponse(c echo.Context, items interface{}, message string, totalRecord, totalRecordPerPage, totalRecordSearch, totalPage int64, currentPage int) error {
	rowNumberStart := (currentPage-1)*int(totalRecordPerPage) + 1
	var rowNumberEnd int
	if currentPage == int(totalPage) {
		rowNumberEnd = int(totalRecord)
	} else {
		rowNumberEnd = currentPage * int(totalRecordPerPage)
	}

	response := &ResponsePagination{
		Records:            items,
		TotalRecord:        totalRecord,
		TotalRecordPerPage: totalRecordPerPage,
		TotalRecordSearch:  totalRecordSearch,
		TotalPage:          totalPage,
		CurrentPage:        currentPage,
		RowNumberStart:     rowNumberStart,
		RowNumberEnd:       rowNumberEnd,
	}

	res := r.SetResponse(c, http.StatusOK, http.StatusText(http.StatusOK), message, response, nil)

	return r.SendResponse(res)
}

func (r *ResponseHelper) SendCustomResponse(c echo.Context, httpCode int, message string, data interface{}) error {
	res := r.SetResponse(c, httpCode, http.StatusText(httpCode), message, data, nil)
	return r.SendResponse(res)
}
