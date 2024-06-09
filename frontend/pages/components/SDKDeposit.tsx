import { InputHTMLAttributes, useEffect, useState } from "react"
import { useAccount, useSendTransaction } from "wagmi"

interface CalldataResponse {
    calldata: {
        input: `0x${string}`
        to: `0x${string}`
        gas: `0x${string}`
        gasPrice: `0x${string}`
    }
}

interface SDKDepositProps {
    baseUri: string
}

export default function SDKDeposit(props: SDKDepositProps) {
    const account = useAccount()
    const transactor = useSendTransaction()

    const [slippage, setSlippage] = useState<number>(0)
    const [amount, setAmount] = useState<number>(0)
    const [price, setPrice] = useState<number>(0)

    useEffect(() => {
        fetch(`${props.baseUri}/slippage`, {
            method: "POST",
            body: JSON.stringify({amount: amount.toString()})
        }).then(data => data.json()).then(data => {
            setSlippage(parseFloat(data.slippage))
        }).catch(() => {
            console.error("Failed to fetch slippage, defaulting to zero.")
            setSlippage(0)
        })
    }, [props.baseUri, amount])

    function captureInput(evt: Parameters<NonNullable<InputHTMLAttributes<HTMLInputElement>['onChange']>>[0]) {
        setAmount(parseFloat(evt.target.value) || 0)
    }

    function createDepositTransaction(amount: number) {
        return async (evt: any) => {
            const data = await fetch(`${props.baseUri}/createDepositTx`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    amount: amount.toString(),
                    from: account.address,
                    token: '0x0000000000000000000000000000000000000000'
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
    
    return <div>
        <input type="text" onChange={captureInput} />
        <button onClick={createDepositTransaction(amount)}>Create deposit transaction for {amount}</button>
    </div>
}