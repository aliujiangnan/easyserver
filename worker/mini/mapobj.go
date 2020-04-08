package main

import (

)

type MapObj struct{
    mapid int
    allPlayers map[int]*Player
    allTeams map[int]*Team
}

func (this *MapObj)GetPlayerNum() int{
    return len(this.allPlayers)
}
func (this *MapObj)AddPlayer(player *Player) {
    this.allPlayers[player.userID] = player
}
func (this *MapObj)GetPlayer(userID int) *Player{
        return this.allPlayers[userID]
    }
func (this *MapObj)DelPlayer(userID int) bool{
    if this.allPlayers[userID] != nil {
        delete(this.allPlayers, userID)
        ForeachTeam(func (team *Team) {
            team.DelPlayer(userID)
        })
        return true
    }
    return false
}
func (this *MapObj)Foreach(fn func(player *Player)) bool{
    for _, playe := range this.allPlayers{
        fn(playe)
    }
    return true
}
func (this *MapObj)AllocPlayerID() int{
    for i := 1;i <= 100; i+=1 {
        if this.allPlayers[i] == nil{
            return i
        }
    }
    return 1
}
func (this *MapObj)GetTeamNum() int{
    return len(this.allTeams)
}
func (this *MapObj)AddTeam(team *Team) {
    this.allTeams[team.teamID] = team
}
func (this *MapObj)GetTeam(teamID int) *Team{
    return this.allTeams[teamID]
}
func (this *MapObj)DelTeam(teamID int) bool{
    if this.allTeams[teamID] != nil {
        delete(this.allTeams,teamID)
        return true
    }
    return false
}
func (this *MapObj)MatchTeam(player *Player, teamType int) *Team{
    var retTeam *Team
    for key, team := range this.allTeams{
        if key >= 0 && team.GetMemberNum() < MAX_NUM_EACH_TEAM && team.selfType == teamType && team.hasRobot == false {
            retTeam = team
        }

    }
    if retTeam == nil {
        retTeam = NewTeam(this.AllocTeamID())
        retTeam.selfType = teamType
    }
    player.team = retTeam
    retTeam.AddPlayer(player)
    this.AddTeam(retTeam)
    AddTeam(retTeam)
    return retTeam
}
func (this *MapObj)AllocTeamID() int{
    for i := 1;i <= 100; i+=1 {
        if this.allTeams[i] == nil{
            return i
        }
        return 1
    }

    return 1
}

var gID = 0
func allocID() int{
    gID += 1
    return gID
}

var gAllMaps = make(map[int]*MapObj)

func AllocMap () *MapObj{
    var retMap *MapObj
    for key, mapObj := range gAllMaps{
        if key >= 0 && mapObj.GetPlayerNum() < MAX_NUM_EACH_MAP {
            retMap = mapObj
            break
        }
            
    }
    if retMap ==  nil{
        retMap = &MapObj{allocID(), make(map[int]*Player), make(map[int]*Team)}
        gAllMaps[retMap.mapid] = retMap
    }
    return retMap
}
