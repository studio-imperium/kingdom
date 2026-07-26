package packets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"server/engine"
	"strconv"

	"github.com/gorilla/websocket"
)

type Client struct {
	id              uint32
	admin           bool
	username        string
	conn            *websocket.Conn
	send            chan []byte
	character       *engine.Character
	instance        *engine.Engine
	simulation      *engine.Engine
	discoveredCells map[uint16]*engine.Cell
}

func (client *Client) destroy() {
	fmt.Println("Client destroyed: ", client.id)
	client.send <- []byte{11}
}

var guests int = 0

func CreateClient(conn *websocket.Conn) {
	guests += 1
	client := Client{
		id:              0,
		admin:           true,
		username:        "Guest_" + strconv.Itoa(guests),
		conn:            conn,
		send:            make(chan []byte, 64),
		character:       nil,
		simulation:      engine.CreateEngine(),
		discoveredCells: make(map[uint16]*engine.Cell),
	}
	go client.recievePackets()
	go client.writePackets()
}

func (client *Client) writePackets() {
	conn := client.conn
	for {
		message, _ := <-client.send

		if message[0] == 11 {
			for _, recipient := range Clients {
				recipient.SendMessage(0, "System", client.username + " has died!")
			}
			client.conn.Close()
			client.instance.RemoveCharacter(client.id, true)
			delete(Clients, client.id)
			break
		}

		err := conn.WriteMessage(websocket.BinaryMessage, message)

		if err != nil {
			return
		}
	}
}

func (client *Client) recievePackets() {
	conn := client.conn
	defer client.destroy()

	for {
		messageType, r, err := conn.NextReader()

		if err != nil || messageType != websocket.BinaryMessage {
			return
		}
		HandlePacket(client, r)
	}
}

func (client *Client) sendToNearby(payload []byte, includeSelf bool) {
	client.simulation.ForEachCharacter(func(id uint32, _ *engine.Character) {
		if !includeSelf && id == client.id {
			return
		}
		if target, ok := Clients[id]; ok {
			target.send <- payload
		}
	})
}

func (client *Client) characterAttack(x float32, y float32, targetX float32, targetY float32, angle uint16) {
	idx := client.instance.GetHand(client.id)
	item_id := client.instance.GetSlot(client.id, idx)
	item := engine.GetItemData(item_id)

	cooldown := client.character.AttackCooldown
	data := new(bytes.Buffer)
	data.WriteByte(uint8(ALLY_ATTACK))

	if item.OnUse != "" {
		client.instance.UseItem(client.id, item.OnUse)
		return
	}

	if cooldown < -0.1 {
		client.character.AttackCounter = 0
	}
	if cooldown <= 0.1 {
		counter := client.character.AttackCounter

		if item.Attacks != nil {
			attack := item.Attacks[int(counter)%len(item.Attacks)]
			animation := attack.Animation
			projectiles := attack.Projectiles
			bombs := attack.Bombs

			// can make this longer/shorter here
			reload := attack.Reload / client.character.Reload

			client.character.AttackCooldown = reload

			binary.Write(data, binary.LittleEndian, client.id)
			data.WriteByte(animation)
			binary.Write(data, binary.LittleEndian, uint16(reload*1000))

			binary.Write(data, binary.LittleEndian, uint16(len(projectiles)))

			for _, projectile := range projectiles {
				baseDamage := engine.GetProjectileData(projectile.ID).Damage

				damage := float32(baseDamage) * client.character.Power
				// we can freely scale damage up/down here based on whatever we want

				id := client.instance.CreateProjectile(projectile.ID, projectile.X+x, projectile.Y+y, (projectile.Angle+angle)%360, false, damage)
				proj := client.instance.Projectiles[id]
				packet := proj.Pack()

				binary.Write(data, binary.LittleEndian, id)
				data.Write(packet)

				client.simulation.AddProjectile(id, proj)
			}

			binary.Write(data, binary.LittleEndian, uint16(len(bombs)))

			for _, bomb := range bombs {
				baseDamage := engine.GetBombData(bomb.ID).Damage
				damage := baseDamage
				timer := engine.GetBombData(bomb.ID).Airtime
				// we can freely scale damage up/down here based on whatever we want
				// and scale timer

				id := client.instance.CreateBomb(bomb.ID, targetX, targetY, client.character, false, damage, timer)
				bomb := client.instance.Bombs[id]
				packet := bomb.Pack()

				binary.Write(data, binary.LittleEndian, id)
				data.Write(packet)

				client.simulation.AddBomb(id, bomb)
			}

			client.character.AttackCounter += 1
			client.character.AttackCounter %= uint8(len(item.Attacks))
			client.sendToNearby(data.Bytes(), false)
		}
	}
}

func (client *Client) sendCharacter() {
	data := client.instance.PackCharacter(client.id, byte(HANDSHAKE))
	client.send <- data
}
