package server

import "github.com/google/btree"

func delCommand(client *Client) {
	if len(client.args) < 2 {
		client.ReplyAritryError()
		return
	}
	count := 0
	for i := 1; i < len(client.args); i++ {
		if _, ok := client.server.DelKey(client.args[i]); ok {
			count++
		}
	}
	client.ReplyInt(count)
}

func renameCommand(client *Client) {
	if len(client.args) != 3 {
		client.ReplyAritryError()
		return
	}
	key, ok := client.server.GetKey(client.args[1])
	if !ok {
		client.ReplyError("no such key")
		return
	}
	client.server.DelKey(client.args[1])
	client.server.SetKey(client.args[2], key)
	client.ReplyString("OK")
}

func renamenxCommand(client *Client) {
	if len(client.args) != 3 {
		client.ReplyAritryError()
		return
	}
	key, ok := client.server.GetKey(client.args[1])
	if !ok {
		client.ReplyError("no such key")
		return
	}
	_, ok = client.server.GetKey(client.args[2])
	if ok {
		client.ReplyInt(0)
		return
	}
	client.server.DelKey(client.args[1])
	client.server.SetKey(client.args[2], key)
	client.ReplyInt(1)
}

func keysCommand(client *Client) {
	if len(client.args) != 2 {
		client.ReplyAritryError()
		return
	}
	var keys []string
	pattern := parsePattern(client.args[1])
	if pattern.All {
		client.server.keys.Ascend(
			func(item btree.Item) bool {
				keys = append(keys, item.(*Key).Name)
				return true
			},
		)
	} else if !pattern.Glob {
		item := client.server.keys.Get(&Key{Name: pattern.Value})
		if item != nil {
			keys = append(keys, item.(*Key).Name)
		}
	} else if pattern.GreaterOrEqual != "" {
		client.server.keys.AscendRange(
			&Key{Name: pattern.GreaterOrEqual},
			&Key{Name: pattern.LessThan},
			func(item btree.Item) bool {
				if pattern.Match(item.(*Key).Name) {
					keys = append(keys, item.(*Key).Name)
				}
				return true
			},
		)
	} else {
		client.server.keys.Ascend(
			func(item btree.Item) bool {
				if pattern.Match(item.(*Key).Name) {
					keys = append(keys, item.(*Key).Name)
				}
				return true
			},
		)
	}
	client.ReplyMultiBulkLen(len(keys))
	for _, key := range keys {
		client.ReplyBulk(key)
	}
}

func typeCommand(client *Client) {
	if len(client.args) != 2 {
		client.ReplyAritryError()
		return
	}
	key, ok := client.server.GetKey(client.args[1])
	if !ok {
		client.ReplyString("none")
		return
	}
	switch key.(type) {
	default:
		client.ReplyString("unknown") // should not be reached
	case string:
		client.ReplyString("string")
	}
}

func randomkeyCommand(client *Client) {
	if len(client.args) != 1 {
		client.ReplyAritryError()
		return
	}
	item := client.server.keys.Random()
	if item == nil {
		client.ReplyNull()
		return
	}
	client.ReplyString(item.(*Key).Name)
}