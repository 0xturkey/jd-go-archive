import axios from "axios";
import { HDNodeWallet, Mnemonic, JsonRpcProvider } from "ethers";
import { config } from "dotenv";

config();

const mnemonicPhrase = process.env.MNEMONIC;

if (!mnemonicPhrase) {
    throw new Error("MNEMONIC is not set");
}

export const sepoliaEndpoint = "https://rpc-sepolia.rockx.com";

export const sepoliaProvider = new JsonRpcProvider(sepoliaEndpoint);

const mnemonic = Mnemonic.fromPhrase(mnemonicPhrase);

export function getWallet(index: number): HDNodeWallet {
    return HDNodeWallet.fromMnemonic(mnemonic, `m/44'/60'/0'/0/${index}`);
}

export const defaultWallet = getWallet(0);

export const client = axios.create({
    baseURL: "http://localhost:3000",
    headers: {
        "Content-Type": "application/json",
    },
});
