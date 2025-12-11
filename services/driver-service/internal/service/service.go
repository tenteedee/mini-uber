package service

import (
	math "math/rand/v2"
	"sync"

	"github.com/mmcloughlin/geohash"

	"github.com/tenteedee/mini-uber/services/driver-service/internal/domain"
	"github.com/tenteedee/mini-uber/services/driver-service/utils"
	pb "github.com/tenteedee/mini-uber/shared/proto/driver"
	sharedUtils "github.com/tenteedee/mini-uber/shared/util"
)

type Service struct {
	drivers []*domain.DriverInMap
	mu      sync.RWMutex
}

func NewService() *Service {
	return &Service{
		drivers: make([]*domain.DriverInMap, 0),
	}
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIndex := math.IntN(len(utils.PredefinedRoutes))
	randomRoute := utils.PredefinedRoutes[randomIndex]

	randomPlate := utils.GenerateRandomPlate()
	randomAvatar := sharedUtils.GetRandomAvatar(randomIndex)

	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver := &pb.Driver{
		Id:             driverId,
		Name:           "Driver 1",
		PackageSlug:    packageSlug,
		ProfilePicture: randomAvatar,
		CarPlate:       randomPlate,
		Geohash:        geohash,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
	}

	s.drivers = append(s.drivers, &domain.DriverInMap{
		Driver: driver,
	})

	return driver, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, driver := range s.drivers {
		if driver.Driver.Id == driverId {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
			break
		}
	}
}

func (s *Service) FindAvailableDrivers(packageType string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mathcingDrivers []string

	for _, drivers := range s.drivers {
		if drivers.Driver.PackageSlug == packageType {
			mathcingDrivers = append(mathcingDrivers, drivers.Driver.Id)
		}
	}

	if len(mathcingDrivers) == 0 {
		return []string{}
	}

	return mathcingDrivers
}
