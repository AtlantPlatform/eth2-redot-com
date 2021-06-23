package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var epochTemplate = template.Must(template.New("epoch").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/epoch.html"))
var epochNotFoundTemplate = template.Must(template.New("epochnotfound").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/epochnotfound.html"))

// Epoch will show the epoch using a go template
func Epoch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	epochString := strings.Replace(vars["epoch"], "0x", "", -1)

	data := InitPageData(w, r, "epochs", "/epochs", "Epoch")
	data.HeaderAd = true

	epoch, err := strconv.ParseUint(epochString, 10, 64)
	data.Meta.Title = fmt.Sprintf("Epoch %v - Ethereum 2.0 Beacon Chain Explorer | Redot", epochString)
	data.Meta.Description = fmt.Sprintf("Epoch %v - Ethereum 2.0 Beacon Chain (Phase 0) Block Chain Explorer provides easy way to search for Ethereum 2 epochs.", epochString)
	data.Meta.Path = data.Meta.Webroot + "/epoch/" + epochString

	if err != nil {
		logger.Errorf("error parsing epoch index %v: %v", epochString, err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	epochPageData := types.EpochPageData{}

	err = db.DB.Get(&epochPageData, `SELECT epoch, 
											    blockscount, 
											    proposerslashingscount, 
											    attesterslashingscount, 
											    attestationscount, 
											    depositscount, 
											    voluntaryexitscount, 
											    validatorscount, 
											    averagevalidatorbalance, 
											    finalized,
											    eligibleether,
											    globalparticipationrate,
											    votedether
										FROM epochs 
										WHERE epoch = $1`, epoch)
	if err != nil {
		//logger.Errorf("error getting epoch data: %v", err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	err = db.DB.Select(&epochPageData.Blocks, `SELECT blocks.slot, 
											    blocks.proposer, 
											    blocks.blockroot, 
											    blocks.parentroot, 
											    blocks.attestationscount, 
											    blocks.depositscount, 
											    blocks.voluntaryexitscount, 
											    blocks.proposerslashingscount, 
											    blocks.attesterslashingscount,
       										blocks.status
										FROM blocks 
										WHERE epoch = $1
										ORDER BY blocks.slot DESC`, epoch)

	if err != nil {
		logger.Errorf("error epoch blocks data: %v", err)
		err = epochNotFoundTemplate.ExecuteTemplate(w, "layout", data)

		if err != nil {
			logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
			http.Error(w, "Internal server error", 503)
			return
		}
		return
	}

	for _, block := range epochPageData.Blocks {
		block.Ts = utils.SlotToTime(block.Slot)

		switch block.Status {
		case 0:
			epochPageData.ScheduledCount += 1
		case 1:
			epochPageData.ProposedCount += 1
		case 2:
			epochPageData.MissedCount += 1
		case 3:
			epochPageData.OrphanedCount += 1
		}
	}

	epochPageData.Ts = utils.EpochToTime(epochPageData.Epoch)

	err = db.DB.Get(&epochPageData.NextEpoch, "SELECT epoch FROM epochs WHERE epoch > $1 ORDER BY epoch LIMIT 1", epochPageData.Epoch)
	if err != nil {
		logger.Errorf("error retrieving next epoch for epoch %v: %v", epochPageData.Epoch, err)
		epochPageData.NextEpoch = 0
	}
	err = db.DB.Get(&epochPageData.PreviousEpoch, "SELECT epoch FROM epochs WHERE epoch < $1 ORDER BY epoch DESC LIMIT 1", epochPageData.Epoch)
	if err != nil {
		logger.Errorf("error retrieving previous epoch for epoch %v: %v", epochPageData.Epoch, err)
		epochPageData.PreviousEpoch = 0
	}

	data.Data = epochPageData

	if utils.IsApiRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data.Data)
	} else {
		err = epochTemplate.ExecuteTemplate(w, "layout", data)
	}

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
