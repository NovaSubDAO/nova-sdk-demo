import { useEffect } from "react"

const STABLECOINS = [
    {
        icon: "https://assets.coingecko.com/coins/images/9956/small/dai-multi-collateral-mcd.png",
        name: "DAI",
        decimals: 18
    },
    {
        icon: "https://assets.coingecko.com/coins/images/6319/small/USD_Coin_icon.png",
        name: "USDC",
        decimals: 6
    },
    {
        icon: "https://assets.coingecko.com/coins/images/325/small/Tether-logo.png",
        name: "USDT",
        decimals: 6
    }
]

export interface Stablecoin {
    symbol: string,
    address: string,
    decimals: number
}

interface SelectStablecoinProps {
    onSelect: (stablecoin: Stablecoin) => void,
    supportedStablecoins: Stablecoin[]
}

export default function SelectStablecoin(props: SelectStablecoinProps) {
    useEffect(() => {
        console.log("select stablecoins effect")
        if (!props.supportedStablecoins) return
        props.onSelect(props.supportedStablecoins[0])
    }, [props.supportedStablecoins])

    return (
        <div>
            <div style={({display: "flex", flexBasis: "row", justifyContent: "space-around"})}>
                {STABLECOINS.filter(coin => (props.supportedStablecoins.map(c => c.symbol) || []).includes(coin.name)).map((coin) => {
                    return (
                        <div key={coin.name} onClick={() => props.onSelect(props.supportedStablecoins.find(c => c.symbol === coin.name)!)}>
                            <img src={coin.icon} alt={coin.name} height="22px"/>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}