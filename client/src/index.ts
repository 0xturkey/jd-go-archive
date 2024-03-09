import { getJwtToken } from "./login";
import { sepoliaProvider, defaultWallet, getWallet, client, sepoliaEndpoint } from "./utils";

async function main() {
    console.log("defaultWallet: ", defaultWallet.address);

    const balance = await sepoliaProvider.getBalance(defaultWallet.address);

    console.log("balance: ", balance.toString());

    const token = await getJwtToken();

    const recipient = getWallet(1).address;

    const populatedTransaction = await defaultWallet.connect(sepoliaProvider).populateTransaction({
        to: recipient,
        value: 100_000_000,
    });

    const signedTx = await defaultWallet.signTransaction(populatedTransaction);

    console.log("signedTx: ", signedTx);

    console.log({ token });

    const result = await client.post("/tasks", {
        wallet_address: defaultWallet.address,
        rpc_url: sepoliaEndpoint,
        encoded_transaction: signedTx,
        scheduled_at: new Date().toISOString(),
        task_type: "ETH_TX",
    }, {
        headers: {
            Authorization: token,
        }
    });

    console.log(result.data);
}

main().catch((error: any) => {
    console.error(error);
    console.error("exiting with error");
    process.exit(1);
});
