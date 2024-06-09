import { useEffect, useState } from "react"
import { useAccount, useSendTransaction } from "wagmi"
import SDKDeposit from "./SDKDeposit"
import SDKWithdraw from "./SDKWithdraw"


export default function SDKApp(props: any) {
    const account = useAccount()
    const transactor = useSendTransaction()

    const [baseUri, setBaseUri] = useState<string>('http://localhost:8000/main')
    const [position, setPosition] = useState<number>(0)
    const [price, setPrice] = useState<number>(0)

    useEffect(() => {
        setBaseUri(`http://localhost:8000/${account.chainId === 1? 'main': 'opt'}`)
    }, [account.chainId])

    useEffect(() => {
        if (!account.address) return
        fetch(`${baseUri}/position`, {
            method: "POST",
            body: JSON.stringify({address: account.address})
        }).then(data => data.json()).then(data => {
            console.log("position", data)
        })
        fetch(`${baseUri}/price`).then(data => data.json()).then(data => {
            console.log("price", data)
            setPrice(parseFloat(data.price))
        })
    }, [baseUri, account.address])


    return (
        <div>
            <h2>SDKApp</h2>
            connected to {account.chainId === 1? "mainnet": "optimism"}
            <p>price {price}</p>
            <p>position {position}</p>
            <span>{props.address}</span>
            <SDKDeposit baseUri={baseUri} />
            <SDKWithdraw baseUri={baseUri} />
        </div>
    )
}