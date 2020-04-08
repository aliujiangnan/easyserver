package main

import (
    "log"

    "easyserver/lib/util"
    "easyserver/lib/worker"
)

type Player struct{
    sessionid int
    mapObj *MapObj
    team *Team
    playerID int
    userID int
    userName string
    nickName string
    avatar string
    gender int
    score int
    exp int
    level int
    online bool
    isPlayer bool
    ready bool
}

func (this *Player)GetID() int{
    return this.sessionid
}
func (this *Player)SendMsg(name string, data map[string]interface{}) {
    worker.SendMsg(this.sessionid, "cmd:0\n"+util.JsonToString(map[string]interface{}{"n":name,"d":data}))
}
func (this *Player)BroadcastInMap(name string, data map[string]interface{}) {
    if this.mapObj != nil{
        this.mapObj.Foreach(func (player *Player) {
            player.SendMsg(name, data)
         })
    }
}
func (this *Player)BroadcastInTeam(name string, data map[string]interface{}) {
    if this.team != nil{
        this.team.Foreach(func (player *Player) {
            player.SendMsg(name, data)
        })
    }
}
func (this *Player)NotifyOtherInTeam(name string, data map[string]interface{}) {
    if this.team != nil{
        this.team.Foreach(func (player *Player) {
            if player.GetID() != this.sessionid {
                player.SendMsg(name, data)
            }
        })
    }
}

var gAllPlayers = make(map[int]*Player)
var gAllSessions = make(map[int]int)

func NewPlayer(id int) *Player{
    return  &Player{id,nil,nil,0,0,"","","",0,0,0,0,true,true,false}
}
func AddPlayer (player *Player) {
    gAllPlayers[player.userID] = player
    gAllSessions[player.GetID()] = player.userID
}
func GetPlayer (id int) *Player{
    return gAllPlayers[id]
}

func GetPlayerBySessonId (id int) *Player{
    uid := gAllSessions[id]
    return gAllPlayers[uid]
}
func DelPlayer (userID int) bool{
    if gAllPlayers[userID] != nil {
        sessionID := gAllPlayers[userID].GetID()
        delete(gAllPlayers, userID)
        if gAllSessions[sessionID] != 0 {
            delete(gAllSessions, sessionID)
        }
        return true
    }
    return false
}
func ForeachPlayer (fn func(player *Player)) {
    for _, playe := range gAllPlayers{
        fn(playe)
    }
}
