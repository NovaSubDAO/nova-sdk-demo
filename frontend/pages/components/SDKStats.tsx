import { InputHTMLAttributes, useEffect, useState } from "react"
import { useAccount, useSendTransaction } from "wagmi"
import SelectStablecoin from "./SelectStablecoin"

interface SDKStatsProps {
    baseUri: string
}

export default function SDKStats(props: SDKStatsProps) {
    const account = useAccount()
    const transactor = useSendTransaction()

    const [userPosition, setUserPosition] = useState<number>(0)
    const [price, setPrice] = useState<number>(0)
    const [canonicalPrice, setCanonicalPrice] = useState<number>(0)

    useEffect(() => {
        if (!account.address || !props.baseUri) return
        fetch(`${props.baseUri}/position`, {
            method: "POST",
            body: JSON.stringify({
                address: account.address,
                stablecoin: "USDC"
            })
        }).then(data => data.json()).then(data => {
            console.log("position", data)
            setUserPosition(data.position)
        })
        fetch(`${props.baseUri}/price`).then(data => data.json()).then(data => {
            console.log("price", data)
            setPrice(parseFloat(data.price))
        })
        fetch(`${props.baseUri}/canonicalPrice`).then(data => data.json()).then(data => {
            console.log("canonicalPrice", data)
            setCanonicalPrice(parseFloat(data.price))
        })
    }, [props.baseUri, account.address])
    
    return <div>
        <h3>Stats</h3>
        <p>Position value: {userPosition} USDC</p>
        <p>Local sDAI price: {price}</p>
        <p>Canonical sDAI price (on Mainnet): {canonicalPrice}</p>
        <p>Difference between canonical price and local price is {Math.abs(price - canonicalPrice) * 100 / canonicalPrice}%</p>
    </div>
}