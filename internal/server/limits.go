package server
type Limits struct{Tier string;Description string}
func LimitsFor(tier string)Limits{if tier=="pro"{return Limits{Tier:"pro",Description:"Pro tier"}};return Limits{Tier:"free",Description:"Free tier"}}
func(l Limits)IsPro()bool{return l.Tier=="pro"}
