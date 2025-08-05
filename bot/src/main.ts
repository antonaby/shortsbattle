import * as dotenv from 'dotenv';
import { Bot } from "grammy";

dotenv.config();

const bot = new Bot(process.env.TELEGRAM_BOT_KEY);

bot.command("start", (ctx) => ctx.reply("Welcome! Up and running."));

bot.command("game", (ctx) => {
  return ctx.reply("Welcome to the Game!", {
    reply_markup: {
      inline_keyboard: [[
        {
          text: "Open Game App",
          web_app: {
            url: process.env.TELEGRAM_BOT_WEB_APP_URL
          },
        },
      ]],
    },
  });
});

bot.on("message", (ctx) => {
  ctx.reply("Got another message!");
});

bot.start();