import { InputHTMLAttributes, useEffect, useState } from "react"
import { useAccount, useSendTransaction } from "wagmi"
import SelectStablecoin from "./SelectStablecoin"

interface CalldataResponse {
    calldata: {
        input: `0x${string}`
        to: `0x${string}`
        gas: `0x${string}`
        gasPrice: `0x${string}`
    }
}

interface SDKWithdrawProps {
    baseUri: string
}

export default function SDKWithdraw(props: SDKWithdrawProps) {
    const account = useAccount()
    const transactor = useSendTransaction()

    const [selectedCoin, setSelectedCoin] = useState<string>("")
    const [slippage, setSlippage] = useState<number>(0)
    const [amount, setAmount] = useState<number>(0)
    const [price, setPrice] = useState<number>(0)

    useEffect(() => {
        fetch(`${props.baseUri}/slippage`, {
            method: "POST",
            body: JSON.stringify({
                amount: amount.toString(),
                address: account.address,
                // FIXME: should be outputToken
                inputToken: selectedCoin
            })
        }).then(data => data.json()).then(data => {
            setSlippage(parseFloat(data.slippage))
        }).catch(() => {
            console.error("Failed to fetch slippage, defaulting to zero.")
            setSlippage(0)
        })
    }, [props.baseUri, amount, selectedCoin])

    function captureInput(evt: Parameters<NonNullable<InputHTMLAttributes<HTMLInputElement>['onChange']>>[0]) {
        setAmount(parseFloat(evt.target.value) || 0)
    }

    function createWithdrawTransaction(amount: number) {
        return async (evt: any) => {
            const data = await fetch(`${props.baseUri}/createWithdrawTx`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    amount: amount.toString(),
                    from: account.address,
                    token: selectedCoin
                })
            }).then(res => res.json()).then(data => ({calldata: JSON.parse(data.calldata)})) as CalldataResponse

            console.log(data)

            const tx = transactor.sendTransaction({
                to: data.calldata.to,
                data: data.calldata.input,
                value: BigInt(0),
                chainId: account.chainId,
                gas: BigInt(data.calldata.gas),
                gasPrice: BigInt(data.calldata.gasPrice)
            })
        }
    }
    
    return <div className="card" style={({border: "1px solid #f0c", borderRadius: "1em", padding: "1em"})}>
        <h3>Withdraw</h3>
        <SelectStablecoin onSelect={setSelectedCoin} />
        <input type="text" onChange={captureInput} />
        <button onClick={createWithdrawTransaction(amount)}>Withdraw {amount} sDAI receiving {selectedCoin}</button>
        <p>Slippage: {slippage}</p>
    </div>
}