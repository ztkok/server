package main

func (p *RoomUserMsgProc) RPC_PutSkill(skillID uint32) {
	p.user.Debug("RPC_PutSkill skillID:", skillID)

	p.user.putInitiveSkill(skillID)
}
