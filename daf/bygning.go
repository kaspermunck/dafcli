package daf

import "fmt"

// Bygning is a single BBR_Bygning node.
//
// Field names with Danish characters (ø, å, æ) trigger HotChocolate parse
// errors (HC0011) on this Datafordeler stack — fields like
// byg026Opførelsesår and byg041BebyggetAreal are omitted from v1. Re-add
// once the encoding is sorted.
type Bygning struct {
	IDLokalId      string `json:"id_lokalId"`
	Status         string `json:"status"`
	Bygningsnummer int    `json:"byg007Bygningsnummer"`
	Anvendelse     string `json:"byg021BygningensAnvendelse"`
	Husnummer      string `json:"husnummer"`
}

type bygningData struct {
	BBRBygning struct {
		Nodes []Bygning `json:"nodes"`
	} `json:"BBR_Bygning"`
}

// BygningerByHusnummer returns all buildings registered against an access-
// address UUID (DAR Husnummer.id_lokalId, same as DAWA's adgangsadresseid).
func (c *Client) BygningerByHusnummer(husnummerUUID string, limit int) ([]Bygning, error) {
	if limit <= 0 {
		limit = 20
	}
	ts := NowTimestamp()
	query := fmt.Sprintf(`{
		BBR_Bygning(registreringstid:%q, virkningstid:%q, where:{husnummer:{eq:%q}}, first:%d) {
			nodes {
				id_lokalId
				status
				byg007Bygningsnummer
				byg021BygningensAnvendelse
				husnummer
			}
		}
	}`, ts, ts, husnummerUUID, limit)

	raw, err := c.QueryRaw("BBR", query)
	if err != nil {
		return nil, err
	}
	data, err := decodeGraphQL[bygningData](raw)
	if err != nil {
		return nil, err
	}
	return data.BBRBygning.Nodes, nil
}

// BBRAnvendelseLabel resolves a 3-digit BBR anvendelseskode to a Danish label.
// Partial coverage — extend as new codes appear in your queries.
func BBRAnvendelseLabel(code string) string {
	switch code {
	case "110":
		return "Stuehus til landbrugsejendom"
	case "120":
		return "Fritliggende enfamiliehus"
	case "130":
		return "Række-/kæde-/dobbelthus"
	case "140":
		return "Etagebolig"
	case "150":
		return "Kollegium"
	case "160":
		return "Boligbygning til døgninstitution"
	case "190":
		return "Anden bygning til helårsbolig"
	case "210":
		return "Bygning til erhvervsmæssig produktion"
	case "220":
		return "Bygning til energiproduktion"
	case "230":
		return "Bygning til transport"
	case "310":
		return "Hotel/restaurant/forsamlingshus"
	case "320":
		return "Bygning til kontor/handel/lager (samlet)"
	case "321":
		return "Bygning til kontor"
	case "322":
		return "Bygning til detailhandel"
	case "323":
		return "Lagerbygning"
	case "324":
		return "Bygning til offentlig administration"
	case "390":
		return "Anden bygning til kontor/handel/lager"
	case "410":
		return "Bygning til biograf/teater/koncert"
	case "411":
		return "Bygning til museum/bibliotek"
	case "412":
		return "Bygning til kirke/religion"
	case "413":
		return "Bygning til skole/uddannelse"
	case "414":
		return "Bygning til universitet/forskning"
	case "415":
		return "Bygning til hospital/sygehus"
	case "416":
		return "Bygning til sundhedscenter/lægehus"
	case "417":
		return "Bygning til daginstitution"
	case "420":
		return "Bygning til kultur/sport/fritid (samlet)"
	case "421":
		return "Bygning til daginstitution"
	case "422":
		return "Bygning til skole"
	case "510":
		return "Sommerhus"
	case "520":
		return "Feriebolig (helårsanvendelse)"
	case "530":
		return "Anden ferie-/fritidsbygning"
	case "910":
		return "Garage til 1-2 køretøjer"
	case "920":
		return "Carport"
	case "930":
		return "Udhus"
	case "940":
		return "Drivhus"
	case "990":
		return "Anden småbygning"
	}
	return ""
}
