package handlers

import (
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

var chartsTemplate = template.Must(template.New("charts").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/charts.html"))
var genericChartTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/genericchart.html"))
var chartsUnavailableTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/chartsunavailable.html"))
var slotVizTemplate = template.Must(template.New("slotViz").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/slotViz.html"))

func getChartMeta(chartVar string) (string, string) {
	switch chartVar {
	case "blocks":
		return "Ethereum 2.0 Blocks Daily Chart | Redot",
			"The easiest way to check out the history of ETH 2.0 daily blocks proposed using block explorer on redot."
	case "validators":
		return "Ethereum 2.0 Active Validators Daily Chart | Redot",
			"The easiest way to check out the history of ETH 2.0 daily active validators using block explorer on redot."
	default:
		return "Ethereum 2.0 (ETH) Charts - Redot",
			"Check out Ethereum 2.0 (ETH) Charts - Ethereum 2.0 Beacon Chain (Phase 0) Block Chain Explorer"
	}
}

func getChartHeader(chartVar string) string {
	switch chartVar {
	case "blocks":
		return "Ethereum 2.0 Blocks Chart"
	case "validators":
		return "Ethereum 2.0 Validators Chart"
	default:
		return "Charts from the Ethereum 2.0 Network"
	}
}

// Charts uses a go template for presenting the page to show charts
func Charts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "stats", "/charts", "Charts")
	chartsPageData := services.LatestChartsPageData()
	if chartsPageData == nil {
		err := chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	data.Data = chartsPageData
	data.Meta.Title = "Ethereum 2.0 (ETH) Charts - Redot"
	data.Meta.Description = "Check out Ethereum 2.0 (ETH) Charts - Ethereum 2.0 Beacon Chain (Phase 0) Block Chain Explorer"

	chartsTemplate = template.Must(template.New("charts").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/charts.html"))
	err := chartsTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// Chart renders a single chart
func Chart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chartVar := vars["chart"]
	switch chartVar {
	case "slotviz":
		SlotViz(w, r)
	default:
		GenericChart(w, r)
	}
}

// GenericChart uses a go template for presenting the page of a generic chart
func GenericChart(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chartVar := vars["chart"]

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "stats", "/charts", "Chart")
	data.Meta.Title, data.Meta.Description = getChartMeta(chartVar)

	chartsPageData := services.LatestChartsPageData()
	if chartsPageData == nil {
		err := chartsUnavailableTemplate.ExecuteTemplate(w, "layout", data)
		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	var chartData *types.GenericChartData
	for _, d := range *chartsPageData {
		if d.Path == chartVar {
			chartData = d.Data
			break
		}
	}
	chartData.H1 = getChartHeader(chartVar)

	if chartData == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	data.Data = chartData

	genericChartTemplate = template.Must(template.New("chart").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/genericchart.html"))
	err := genericChartTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// SlotViz renders a single page with a d3 slot (block) visualisation
func SlotViz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "stats", "/charts", "Charts")

	data.Data = nil
	err := slotVizTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
