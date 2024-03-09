import { SiweMessage } from "siwe";
import { Wallet } from "ethers";

import { defaultWallet, client } from "./utils";

async function signLoginMessage(): Promise<{ message: string; signature: string }> {
    const message = new SiweMessage({
      domain: "jd.io",
      address: defaultWallet.address,
      statement: "hello",
      uri: "https://jd.io",
      version: '1',
      chainId: 1,
    });
    const rawMessage = message.prepareMessage();
  
    const signature = await defaultWallet.signMessage(rawMessage);

    return {
        message: rawMessage,
        signature,
    }
}

async function createUser() {
    const { message, signature } = await signLoginMessage();

    const result = await client.post("/users", {
        primary_address: defaultWallet.address,
        message: message,
        signature: signature,
    });

    console.log(result.data);

    return result.data.token;
}

async function login() {
    const { message, signature } = await signLoginMessage();

    const result = await client.post("/users/login", {
        primary_address: defaultWallet.address.toLowerCase(),
        message: message,
        signature: signature,
    });

    return result.data.token;
}

export async function getJwtToken() {
    try {
        const token = await login();
        return token;
    } catch (error: any) {
        console.log(error?.response?.data?.message);
        if (error?.response?.data?.message === "User not found") {
            const token = await createUser();
            return token;
        }
        
        throw error;
    }
}
