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

interface SDKWithdrawProps {
    baseUri: string
}

export default function SDKWithdraw(props: SDKWithdrawProps) {
    const account = useAccount()
    const transactor = useSendTransaction()
    const { writeContract } = useWriteContract()

    const [selectedCoin, setSelectedCoin] = useState<Stablecoin>({symbol: "", address: "", decimals: 0})
    const [slippage, setSlippage] = useState<number>(0)
    const [amount, setAmount] = useState<number>(0)
    const [price, setPrice] = useState<number>(0)
    const [supportedStablecoins, setSupportedStablecoins] = useState<Stablecoin[]>([])

    useEffect(() => {
        fetch(`${props.baseUri}/supportedStablecoins`).then(data => data.json()).then(data => {
            setSupportedStablecoins(data)
        })
    }, [props.baseUri])

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

    async function setApproval(amount: number) {
        // NOTE: this should be for sDAI
        const vaultAddr = await fetch(`${props.baseUri}/vaultAddress`).then(data => data.json()).then(data => data.address)
        console.log(vaultAddr, selectedCoin.address, TOKEN_ABI, amount * 18)
        writeContract({
            abi: TOKEN_ABI,
            address: props.baseUri.includes("opt")? "0x2218a117083f5B482B0bB821d27056Ba9c04b1D3":"0x83F20F44975D03b1b09e64809B757c47f942BEeA",
            functionName: "approve",
            args: [vaultAddr, parseUnits(amount.toString(), 18)]
        })
    }

    async function createWithdrawTransaction(amount: number) {
        console.log(parseUnits(amount.toString(), 18).toString())
        const data = await fetch(`${props.baseUri}/createWithdrawTx`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                amount: parseUnits(amount.toString(), 18).toString(),
                from: account.address,
                token: selectedCoin.symbol
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
    
    return <div className="card" style={({border: "1px solid #f0c", borderRadius: "1em", padding: "1em"})}>
        <h3>Withdraw</h3>
        {supportedStablecoins && <SelectStablecoin onSelect={setSelectedCoin} supportedStablecoins={supportedStablecoins}/>}
        <input type="text" onChange={captureInput} />
        {selectedCoin && <button onClick={() => setApproval(amount)}>Approve {amount} sDAI</button>}
        {selectedCoin && <button onClick={() => createWithdrawTransaction(amount)}>Withdraw {amount} sDAI receiving {selectedCoin.symbol}</button>}
        <p>Slippage: {slippage.toFixed(2)}%</p>
    </div>
}