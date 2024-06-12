import { useEffect, useState } from "react"
import { useAccount, useSendTransaction } from "wagmi"
import SDKDeposit from "./SDKDeposit"
import SDKWithdraw from "./SDKWithdraw"
import SDKStats from "./SDKStats"

export default function SDKApp(props: any) {
    const account = useAccount()
    const transactor = useSendTransaction()

    const [baseUri, setBaseUri] = useState<string>("")
    const [position, setPosition] = useState<number>(0)
    const [price, setPrice] = useState<number>(0)

    useEffect(() => {
        //setBaseUri(`http://localhost:8000/${account.chainId === 1? 'main': 'opt'}`)
        setBaseUri(`${window.location.protocol}//${window.location.host}/${account.chainId === 1? 'main': 'opt'}`)
    }, [account.chainId])

    return (
        <div>
            <h2>SDKApp</h2>
            connected to {account.chainId === 1? "mainnet": "optimism"}
            <span>{props.address}</span>
            {baseUri && <>
                <SDKStats baseUri={baseUri} />
                <SDKDeposit baseUri={baseUri} />
                <SDKWithdraw baseUri={baseUri} />
            </>}
        </div>
    )
}