import { InputHTMLAttributes, useEffect, useState } from "react"
import { useAccount, useSendTransaction, useWriteContract } from "wagmi"
import SelectStablecoin, { Stablecoin } from "./SelectStablecoin"
import TOKEN_ABI from "../../abi/tokenAbi"
import { formatUnits, parseUnits } from "viem"

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
    const { writeContract } = useWriteContract()

    const [slippage, setSlippage] = useState<number>(0)
    const [amount, setAmount] = useState<number>(0)
    const [selectedCoin, setSelectedCoin] = useState<Stablecoin>({symbol: "", address: "", decimals: 0})
    const [price, setPrice] = useState<number>(0)
    const [supportedStablecoins, setSupportedStablecoins] = useState<Stablecoin[]>([])

    useEffect(() => {
        fetch(`${props.baseUri}/supportedStablecoins`).then(data => data.json()).then(data => {
            setSupportedStablecoins(data)
        })
    }, [props.baseUri])

    useEffect(() => {
        selectedCoin && fetch(`${props.baseUri}/slippage`, {
            method: "POST",
            body: JSON.stringify({
                amount: amount.toString(),
                address: account.address,
                inputToken: selectedCoin.symbol
            })
        }).then(data => data.json()).then(data => {
            console.log(data)
            setSlippage(parseFloat(data.slippage))
        }).catch(() => {
            console.error("Failed to fetch slippage, defaulting to zero.")
            setSlippage(0)
        })
    }, [props.baseUri, amount, selectedCoin])

    function captureInput(evt: Parameters<NonNullable<InputHTMLAttributes<HTMLInputElement>['onChange']>>[0]) {
        setAmount(parseFloat(evt.target.value) || 0)
    }

    async function setApproval(amount: number) {
        const vaultAddr = await fetch(`${props.baseUri}/vaultAddress`).then(data => data.json()).then(data => data.address)
        console.log(vaultAddr, selectedCoin.address, TOKEN_ABI, amount * 10**selectedCoin.decimals)
        writeContract({
            abi: TOKEN_ABI,
            address: selectedCoin.address as any,
            functionName: "approve",
            args: [vaultAddr, parseUnits(amount.toString(), selectedCoin.decimals)]
        })
    }

    function createDepositTransaction(amount: number) {
        return async (evt: any) => {
            const data = await fetch(`${props.baseUri}/createDepositTx`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    amount: formatUnits(parseUnits(amount.toString(), selectedCoin.decimals), 0),
                    from: account.address,
                    token: selectedCoin.symbol
                })
            }).then(res => res.json()).then(data => ({calldata: JSON.parse(data.calldata)})) as CalldataResponse

            console.log(formatUnits(parseUnits(amount.toString(), selectedCoin.decimals), 0))

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
        <h3>Deposit</h3>
        {supportedStablecoins && <SelectStablecoin onSelect={setSelectedCoin} supportedStablecoins={supportedStablecoins}/>}
        <input type="text" onChange={captureInput} />
        {selectedCoin && <button onClick={() => setApproval(amount)}>Approve {amount} {selectedCoin.symbol}</button>}
        {selectedCoin && <button onClick={createDepositTransaction(amount)}>Deposit {amount} {selectedCoin.symbol} for sDAI</button>}
        <p>Slippage: {slippage.toFixed(2)}%</p>
    </div>
}