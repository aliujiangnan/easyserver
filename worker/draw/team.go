package main

import (

)

type Team struct{
    teamID      int
    selfType    int
    creator     int
    allPlayers  map[int]*Player
    allObsers   map[int]*Player
}

func (this *Team)AddPlayer(player *Player){
    this.allPlayers[player.userID] = player
}
func (this *Team)GetPlayer(userID int) *Player{
    return this.allPlayers[userID]
}
func (this *Team)GetPlayerIndex(userID int)int{
    i := 0
    for _, player := range this.allPlayers{
        if player.userID == userID{
            break
        }
        i += 1
    }
    return i
}
func (this *Team)GetPlayerByIndex(index int) *Player{
    i := 0
    for _, player := range this.allPlayers{
        if i == index{
            return player
        }
        i += 1
    }
    return nil
}
func (this *Team)DelPlayer(userID int) bool{
    if this.allPlayers[userID] != nil {
        delete(this.allPlayers, userID)
        return true
    }
    return false
}
func (this *Team)AddObser(player *Player){
    this.allObsers[player.userID] = player
}
func (this *Team)GetObser(userID int) *Player{
    return this.allObsers[userID]
}
func (this *Team)DelObser(userID int) bool{
    if this.allObsers[userID] != nil{
        delete(this.allObsers,userID)
        return true
    }
    return false
}
func (this *Team)GetMemberNum() int{
    return len(this.allPlayers) + len(this.allObsers)
}
func (this *Team)GetPlayerNum() int{
    return len(this.allPlayers)
}
func (this *Team)GetObserNum() int{
    return len(this.allObsers)
}
func (this *Team)GetReadyNum()int{
    i := 0
    for _, player := range this.allPlayers{
        if player.ready{
            i += 1
        }
    }
    return i
}
func (this *Team)Standup(player *Player) bool{
    if (this.GetObserNum() < MAX_OB_EACH_TEAM){
        this.DelPlayer(player.userID)
        this.AddObser(player)
        player.isPlayer = false
        player.ready = false
        return true
    }else{
        return false
    }
}
func (this *Team)Sitdown(player *Player) bool{
    if (this.GetPlayerNum() < MAX_NUM_EACH_TEAM){
        this.DelObser(player.userID)
        this.AddPlayer(player)
        player.isPlayer = true
        player.ready = true
        return true
    }else{
        return false
    }
}

func (this *Team)Foreach(fn func(player *Player)){
    for _, player := range this.allPlayers{
        fn(player)
    }
    for _, player := range this.allObsers{
        fn(player)
    }
}
func (this *Team)ForeachPlayer(fn func(indx int, player *Player)){
    i := 0
    for _, player := range this.allPlayers{
        fn(i, player)
        i += 1
    }
}

func (this *Team)Broadcast(name string, data map[string]interface{}){
    this.Foreach(func(player *Player){
        player.SendMsg(name, data)
    })
}
func (this *Team)getOlPlayerNum() int{
    n := 0
    for _, player := range this.allPlayers{
        if player.online{
            n += 1
        }
    }
    return n
}
func (this *Team)getPainterIndex(count int) int{
    index := -1
    n := 0
    i := 0
    for _, player := range this.allPlayers{
        if player.online{
            if n == count{
                index = i
                break
            }
            n += 1
        }
        i += 1
    }
    return index
}

var gAllTeams = make(map[int]*Team)

func NewTeam(id int) *Team{
    return &Team{id, 0, -1, make(map[int]*Player), make(map[int]*Player)}
}

func AddTeam(team *Team){
    gAllTeams[team.teamID] = team
}

func GetTeam(id int) *Team{
    return gAllTeams[id]
}

func DelTeam(teamID int) bool{
    if gAllTeams[teamID] != nil {
        delete(gAllTeams, teamID)
        return true
    }
    return false
}

func ForeachTeam(fn func(team *Team)){
    for _, tea := range gAllTeams{
        fn(tea)
    }
}