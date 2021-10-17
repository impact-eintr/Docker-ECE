package network

type BridgeNetworkDriver struct {
}

func (d *BridgeNetworkDriver) Name() string {

}

func (d *BridgeNetworkDriver) Create(subnet, name string) (*Network, error) {

}

func (d *BridgeNetworkDriver) Delete(network Network) error {

}

func (d *BridgeNetworkDriver) Connect(net *Network, endpoint *Endpoint) error {

}

func (d *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error { return nil }
