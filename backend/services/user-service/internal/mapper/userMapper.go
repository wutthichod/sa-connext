package mapper

import (
	"strconv"

	"github.com/devfeel/mapper"
	"github.com/jinzhu/copier"

	"github.com/wutthichod/sa-connext/services/user-service/internal/dto"
	"github.com/wutthichod/sa-connext/services/user-service/internal/models"
	pb "github.com/wutthichod/sa-connext/shared/proto/user"
)

func init() {
	// Register DTO → Model
	mapper.Register(&dto.UserDTO{})
	mapper.Register(&dto.ContactDTO{})
	mapper.Register(&dto.EducationDTO{})
}

// FromPbRequest maps Protobuf CreateUserRequest → DTO
func FromPbRequest(req *pb.CreateUserRequest) *dto.UserDTO {
	dtoUser := &dto.UserDTO{
		Username:  req.Username,
		Password:  req.Password,
		JobTitle:  req.JobTitle,
		Interests: req.Interests,
	}

	if req.Contact != nil {
		dtoUser.Contact = dto.ContactDTO{
			Email: req.Contact.Email,
			Phone: req.Contact.Phone,
		}
	}

	if req.Education != nil {
		dtoUser.Education = dto.EducationDTO{
			University: req.Education.University,
			Major:      req.Education.Major,
		}
	}

	return dtoUser
}

// ToUserModel maps DTO → GORM User model (with nested Contact, Education, Interests)
func ToUserModel(dto *dto.UserDTO) *models.User {

	user := models.User{}
	copier.Copy(&user, dto) // dtoUser → user

	// map Interests manually (slice of strings → slice of Interest)
	for _, name := range dto.Interests {
		user.Interests = append(user.Interests, models.Interest{Name: name})
	}

	return &user
}

func ToPbUser(user *models.User) *pb.User {
	interests := make([]string, len(user.Interests))
	for i, interest := range user.Interests {
		interests[i] = interest.Name
	}
	return &pb.User{
		UserId:    strconv.FormatUint(uint64(user.ID), 10),
		Username:  user.Username,
		JobTitle:  user.JobTitle,
		Interests: interests,
		Contact: &pb.Contact{
			Email: user.Contact.Email,
			Phone: user.Contact.Phone,
		},
		Education: &pb.Education{
			University: user.Education.University,
			Major:      user.Education.Major,
		},
	}
}

// FromPbUpdateRequest maps Protobuf UpdateUserRequest → DTO
func FromPbUpdateRequest(req *pb.UpdateUserRequest) *dto.UserDTO {
	dtoUser := &dto.UserDTO{
		Username:  req.Username,
		JobTitle:  req.JobTitle,
		Interests: req.Interests,
	}

	if req.Contact != nil {
		dtoUser.Contact = dto.ContactDTO{
			Email: req.Contact.Email,
			Phone: req.Contact.Phone,
		}
	}

	if req.Education != nil {
		dtoUser.Education = dto.EducationDTO{
			University: req.Education.University,
			Major:      req.Education.Major,
		}
	}

	return dtoUser
}
