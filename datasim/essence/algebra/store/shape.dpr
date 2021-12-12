operations
   =         Initiate {Relation: relation}
//   |          Destroy (RelationRef)
   |   SetPrimaryAttr (EntityRef, AttributeName, AttributeValue)
   | SetSecondaryAttr (EntityRef, EntityRef, AttributeName, AttributeValue)
;

ref
   =   EntityRef (EntityName, EntityID)
//   | RelationRef (RelationId)
;

relation
    = PrimaryWithMany { primary: shape, secondaries: [shape]}
;

shape
    = Entity {Name: Name, Attributes: [attr]}
;

attr
    = Attribute {Name: Name, Type: type}
;

type
    = Int
    | String
    | DateTime
;