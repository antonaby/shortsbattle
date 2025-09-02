import { defineStore } from "pinia";
import {
  Centrifuge,
  Subscription,
  type ConnectionTokenContext,
  type PublicationContext,
  type SubscribedContext,
  type SubscriptionErrorContext,
  type UnsubscribedContext,
} from "centrifuge";
import { ref } from "vue";
import { useUIStore } from "./ui.store";

interface SubMap {
  [channel: string]: Subscription;
}

export const useWSStore = defineStore("ws", () => {
  const uiStore = useUIStore();

  const connected = ref<boolean>(false);
  const subs = {} as SubMap;
  let client: Centrifuge | null = null;

  function updateConnected(value: boolean) {
    connected.value = value;
  }

  function connect(
    tokenProvider: (ctx: ConnectionTokenContext) => Promise<string>
  ) {
    client = new Centrifuge(
      `${import.meta.env.VITE_WS_BASE_URL}/api/v1/games/updates`,
      {
        getToken: tokenProvider, // TODO: check token provider errors
      }
    );

    client.on("connecting", () => updateConnected(false));
    client.on("connected", () => updateConnected(true));
    client.on("disconnected", () => updateConnected(false));

    client.on("error", (ctx) => {
      updateConnected(false);
      uiStore.handleWsError(ctx.type, ctx.error);
    });

    client.connect();
  }

  function disconnect() {
    Object.values(subs).forEach((s) => s.unsubscribe());
    client?.disconnect();
    client = null;
    connected.value = false;
  }

  function subscribe(
    channel: string,
    handlers: {
      subscribed?: (ctx: SubscribedContext) => void;
      unsubscribed?: (ctx: UnsubscribedContext) => void;
      publication?: (ctx: PublicationContext) => void;
      error?: (ctx: SubscriptionErrorContext) => void;
    }
  ): Subscription {
    if (!client) {
      throw new Error("Socket not connected");
    }

    if (subs[channel]) {
      return subs[channel];
    }

    let sub = client.getSubscription(channel);
    if (!sub) {
      sub = client.newSubscription(channel);
    }

    if (handlers.publication) {
      sub.on("publication", handlers.publication);
    }

    if (handlers.subscribed) {
      sub.on("subscribed", handlers.subscribed);
    }
    if (handlers.unsubscribed) {
      sub.on("unsubscribed", handlers.unsubscribed);
    }
    if (handlers.error) {
      sub.on("error", handlers.error);
    }

    sub.subscribe();
    subs[channel] = sub;

    return sub;
  }

  function unsubscribe(channel: string) {
    subs[channel]?.unsubscribe();
    delete subs[channel];
  }

  return { connected, connect, disconnect, subscribe, unsubscribe };
});
