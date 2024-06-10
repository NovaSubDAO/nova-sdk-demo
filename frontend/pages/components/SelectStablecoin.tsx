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

interface SelectStablecoinProps {
    onSelect: (stablecoin: string) => void
}

export default function SelectStablecoin(props: SelectStablecoinProps) {
    useEffect(() => {
        props.onSelect(STABLECOINS[0].name)
    }, [])
    
    return (
        <div>
            <div style={({display: "flex", flexBasis: "row", justifyContent: "space-around"})}>
                {STABLECOINS.map((coin) => {
                    return (
                        <div key={coin.name} onClick={() => props.onSelect(coin.name)}>
                            <img src={coin.icon} alt={coin.name} height="22px"/>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}