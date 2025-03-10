package bot

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	BotToken string
	GuildID  string
	ClientID string
)

var money map[string]int

func Run() {
	money = make(map[string]int)
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal(err)
	}

	discord.AddHandler(newMessage)

	discord.AddHandler(func(session *discordgo.Session, interation *discordgo.InteractionCreate) {
		if interation.Type == discordgo.InteractionApplicationCommand {
			slashCommand(discord, interation)
		} else if interation.Type == discordgo.InteractionMessageComponent {
			buttonClick(discord, interation)
		}
	})

	discord.Open()
	defer discord.Close()

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	fmt.Println("Bot is running")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == discord.State.User.ID {
		return
	}

	switch {
	case strings.HasPrefix(message.Content, "!register"):
		registerCommands(discord, message.GuildID)
	}
}

func buttonClick(discord *discordgo.Session, interation *discordgo.InteractionCreate) {
	switch interation.MessageComponentData().CustomID {
	case "earn:possible":

	case "earn:impossible":

	}
}

func slashCommand(discord *discordgo.Session, interation *discordgo.InteractionCreate) {
	switch interation.ApplicationCommandData().Name {
	case "bank":
		options := interation.ApplicationCommandData().Options

		user := options[0].UserValue(discord)

		_, ok := money[user.ID]

		if !ok {
			if user.ID == interation.Member.User.ID {
				rand.Seed(time.Hour.Milliseconds())
				num := rand.Intn(18) + 2
				money[user.ID] = num
			} else {
				discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "This user does not own a bank account",
					},
				})
				return
			}
		}

		embed := &discordgo.MessageEmbed{
			Author:      &discordgo.MessageEmbedAuthor{},
			Color:       0x00ff00, // Green
			Description: "Daily bank",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Bank Total",
					Value:  "$" + strconv.Itoa(money[user.ID]),
					Inline: true,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://cdn.discordapp.com/avatars/" + user.ID + "/" + user.Avatar + ".webp",
			},
			Timestamp: time.Now().Format(time.RFC3339),
			Title:     user.GlobalName + "'s Bank",
		}

		discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
	case "give":
		options := interation.ApplicationCommandData().Options
		_, ok1 := money[interation.Member.User.ID]

		if !ok1 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You must own a bank account to do this",
				},
			})
			return
		}

		_, ok2 := money[options[0].UserValue(discord).ID]

		if !ok2 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "This user does not own a bank account",
				},
			})
			return
		}

		money[interation.Member.User.ID] -= int(options[1].IntValue())
		money[options[0].UserValue(discord).ID] += int(options[1].IntValue())
		discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Succesfully sent " + options[0].UserValue(discord).GlobalName + " $" + strconv.Itoa(int(options[1].IntValue())),
			},
		})
	case "gamble":
		options := interation.ApplicationCommandData().Options

		multiplier := int(options[0].IntValue())

		_, ok := money[interation.Member.User.ID]

		if !ok {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You don't have a bank account",
				},
			})
			return
		}

		if money[interation.Member.User.ID]-multiplier < 0 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "You don't have enough money to gamble",
				},
			})
			return
		}

		money[interation.Member.User.ID] -= multiplier

		num := rand.Intn(1024)
		if num >= 512 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":white_large_square: Common (1/2 chance) :white_large_square:\nYou got $0 in return",
				},
			})
		} else if num >= 256 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":green_square: Uncommon (1/4 chance) :green_square:\nYou got $" + strconv.Itoa(multiplier/3) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier / 3
		} else if num >= 128 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":blue_square: Rare (1/8 chance) :blue_square:\n You got $" + strconv.Itoa(multiplier/2) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier / 2
		} else if num >= 64 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":purple_square: Epic (1/16 chance) :purple_square:\nYou got $" + strconv.Itoa(multiplier) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier
		} else if num >= 32 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":orange_square: Legendary (1/32 chance) :orange_square:\nYou got $" + strconv.Itoa(multiplier+multiplier/4) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier + multiplier/4
		} else if num >= 16 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":yellow_square: Mythical (1/64 chance) :yellow_square:\nYou got $" + strconv.Itoa(multiplier+multiplier/2) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier + multiplier/2
		} else if num >= 8 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":white_circle: Godlike (1/128 chance) :white_circle:\nYou got $" + strconv.Itoa(multiplier*2) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier * 2
		} else if num >= 4 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":red_circle: Credit (1/256 chance) :red_circle:\nYou got $" + strconv.Itoa(multiplier*4) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier * 4
		} else if num >= 2 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":black_circle: Septillion (1/512 chance) :black_circle:\nYou got $" + strconv.Itoa(multiplier*10) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier * 10
		} else if num >= 1 {
			discord.InteractionRespond(interation.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":o: The One (1/1000 chance) :o:\nYou got $" + strconv.Itoa(multiplier*multiplier) + " in return",
				},
			})
			money[interation.Member.User.ID] += multiplier * multiplier
		}
	case "earn":

	}
}

func registerCommands(discord *discordgo.Session, guild string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "bank",
			Description: "opens bank information",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "a user who owns a bank account",
					Required:    true,
				},
			},
		},
		{
			Name:        "give",
			Description: "gives money to a user",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "a user who owns a bank account",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "money",
					Description: "how much money you want to give",
					Required:    true,
				},
			},
		},
		{
			Name:        "gamble",
			Description: "pay money to gamble and maybe get something in return",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "money",
					Description: "how much money you want to bet",
					Required:    true,
				},
			},
		},
		{
			Name:        "earn",
			Description: "earn money by completing tasks for robots",
		},
	}

	for _, cmd := range commands {
		_, err := discord.ApplicationCommandCreate(discord.State.User.ID, guild, cmd)
		if err != nil {
			fmt.Println("Failed to register commands in guild: " + guild)
		}
	}
}
